package supervisor

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/tgo/captain/aicenter/internal/eino/llm"
)

// PlanExecuteConfig represents configuration for plan-execute mode
type PlanExecuteConfig struct {
	Name          string
	Instruction   string
	Provider      *llm.ProviderConfig
	Tools         []tool.BaseTool
	MaxIterations int
}

// PlanExecuteBuilder builds plan-execute agents
type PlanExecuteBuilder struct {
	llmFactory *llm.Factory
}

// NewPlanExecuteBuilder creates a new PlanExecuteBuilder
func NewPlanExecuteBuilder(llmFactory *llm.Factory) *PlanExecuteBuilder {
	return &PlanExecuteBuilder{
		llmFactory: llmFactory,
	}
}

var executorPrompt = prompt.FromMessages(schema.FString,
	schema.SystemMessage(`You are a task executor. Follow the given plan and execute your tasks carefully.
Execute each planning step by using available tools.
Provide detailed results for each task.
If a tool fails, try alternative approaches or report the issue.`),
	schema.UserMessage(`## OBJECTIVE
{input}
## Given the following plan:
{plan}
## COMPLETED STEPS & RESULTS
{executed_steps}
## Your task is to execute the first step, which is: 
{step}`))

func formatInput(in []adk.Message) string {
	if len(in) == 0 {
		return ""
	}
	return in[0].Content
}

func formatExecutedSteps(in []planexecute.ExecutedStep) string {
	var sb strings.Builder
	for idx, m := range in {
		sb.WriteString(fmt.Sprintf("## %d. Step: %v\n  Result: %v\n\n", idx+1, m.Step, m.Result))
	}
	return sb.String()
}

// Build creates a plan-execute agent
func (b *PlanExecuteBuilder) Build(ctx context.Context, cfg *PlanExecuteConfig) (adk.Agent, error) {
	// Create planner
	plannerModel, err := b.llmFactory.CreateToolCalling(ctx, cfg.Provider)
	if err != nil {
		return nil, fmt.Errorf("create planner model: %w", err)
	}

	planAgent, err := planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
		ToolCallingChatModel: plannerModel,
	})
	if err != nil {
		return nil, fmt.Errorf("create planner: %w", err)
	}

	// Create executor with tools
	executorModel, err := b.llmFactory.CreateToolCalling(ctx, cfg.Provider)
	if err != nil {
		return nil, fmt.Errorf("create executor model: %w", err)
	}

	executeAgent, err := planexecute.NewExecutor(ctx, &planexecute.ExecutorConfig{
		Model: executorModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: cfg.Tools,
			},
		},
		GenInputFn: func(ctx context.Context, in *planexecute.ExecutionContext) ([]adk.Message, error) {
			planContent, err := in.Plan.MarshalJSON()
			if err != nil {
				return nil, err
			}

			firstStep := in.Plan.FirstStep()

			msgs, err := executorPrompt.Format(ctx, map[string]any{
				"input":          formatInput(in.UserInput),
				"plan":           string(planContent),
				"executed_steps": formatExecutedSteps(in.ExecutedSteps),
				"step":           firstStep,
			})
			if err != nil {
				return nil, err
			}

			return msgs, nil
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create executor: %w", err)
	}

	// Create replanner - use ToolCalling model as it implements ChatModel
	replannerModel, err := b.llmFactory.CreateToolCalling(ctx, cfg.Provider)
	if err != nil {
		return nil, fmt.Errorf("create replanner model: %w", err)
	}

	replanAgent, err := planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
		ChatModel: replannerModel,
	})
	if err != nil {
		return nil, fmt.Errorf("create replanner: %w", err)
	}

	// Create plan-execute agent
	maxIterations := cfg.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 20
	}

	return planexecute.New(ctx, &planexecute.Config{
		Planner:       planAgent,
		Executor:      executeAgent,
		Replanner:     replanAgent,
		MaxIterations: maxIterations,
	})
}
