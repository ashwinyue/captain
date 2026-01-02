package supervisor

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

type Runner struct {
	supervisorBuilder *SupervisorBuilder
}

func NewRunner(supervisorBuilder *SupervisorBuilder) *Runner {
	return &Runner{supervisorBuilder: supervisorBuilder}
}

type RunResult struct {
	Content   string
	TotalTime float64
}

// Run executes the team in non-streaming mode
func (r *Runner) Run(ctx context.Context, cfg *SupervisorConfig, query string) (*RunResult, error) {
	return r.RunWithHistory(ctx, cfg, query, nil)
}

// RunWithHistory executes the team with conversation history
func (r *Runner) RunWithHistory(ctx context.Context, cfg *SupervisorConfig, query string, history []*schema.Message) (*RunResult, error) {
	agent, err := r.supervisorBuilder.Build(ctx, cfg)
	if err != nil {
		return nil, err
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		EnableStreaming: false,
		Agent:           agent,
	})

	// Build messages: history + current query
	var messages []*schema.Message
	if len(history) > 0 {
		messages = append(messages, history...)
	}
	messages = append(messages, schema.UserMessage(query))

	iter := runner.Run(ctx, messages)
	var lastMsg adk.Message
	var lastErr error

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			lastErr = event.Err
			continue
		}
		if event.Output != nil {
			lastMsg, _, _ = adk.GetMessage(event)
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return &RunResult{
		Content: lastMsg.Content,
	}, nil
}

// StreamCallback is called for each event during streaming
type StreamCallback func(event *adk.AgentEvent) error

// Stream executes the team in streaming mode and calls the callback for each event
func (r *Runner) Stream(ctx context.Context, cfg *SupervisorConfig, query string, callback StreamCallback) error {
	agent, err := r.supervisorBuilder.Build(ctx, cfg)
	if err != nil {
		return err
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		EnableStreaming: true,
		Agent:           agent,
	})

	iter := runner.Query(ctx, query)
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		// Handle streaming message output
		if event.Output != nil && event.Output.MessageOutput != nil {
			mo := event.Output.MessageOutput
			if mo.IsStreaming && mo.MessageStream != nil {
				// Read from stream and emit message events
				for {
					msg, err := mo.MessageStream.Recv()
					if err != nil {
						break // Stream ended
					}
					// Create a new event with the message content
					streamEvent := &adk.AgentEvent{
						AgentName: event.AgentName,
						Output: &adk.AgentOutput{
							MessageOutput: &adk.MessageVariant{
								Message: msg,
							},
						},
					}
					if err := callback(streamEvent); err != nil {
						return err
					}
				}
				continue
			}
		}

		// Forward non-streaming events directly
		if err := callback(event); err != nil {
			return err
		}
	}
	return nil
}
