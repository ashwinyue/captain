package memory

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	DefaultMaxTokensBeforeSummary     = 4000 // Trigger summarization at 4K tokens
	DefaultMaxTokensForRecentMessages = 1000 // Keep 1K tokens of recent messages
	DefaultMaxMessagesBeforeSummary   = 20   // Or trigger at 20 messages
	DefaultMaxRecentMessages          = 5    // Keep last 5 messages
)

// SummaryPrompt is the system prompt for conversation summarization
const SummaryPrompt = `你是一个对话摘要助手。你的任务是将较长的对话历史压缩成简洁的摘要。

要求：
1. 保留关键信息：用户的主要问题、重要决定、关键结论
2. 保留上下文：用户偏好、已解决的问题、待处理的事项
3. 删除冗余：重复的确认、无关的闲聊、已过时的信息
4. 格式清晰：使用简洁的要点形式

请将以下对话历史压缩成摘要：

{conversation}

请直接输出摘要内容，不要添加额外的解释。`

// SummarizerConfig configures the conversation summarizer
type SummarizerConfig struct {
	// Model used to generate summaries
	Model model.BaseChatModel
	// MaxMessagesBeforeSummary triggers summarization when exceeded
	MaxMessagesBeforeSummary int
	// MaxRecentMessages to keep after summarization
	MaxRecentMessages int
}

// Summarizer compresses conversation history using LLM
type Summarizer struct {
	chain      compose.Runnable[map[string]any, *schema.Message]
	maxMsgs    int
	recentMsgs int
}

// NewSummarizer creates a conversation summarizer
func NewSummarizer(ctx context.Context, cfg *SummarizerConfig) (*Summarizer, error) {
	if cfg.Model == nil {
		return nil, fmt.Errorf("model is required")
	}

	maxMsgs := cfg.MaxMessagesBeforeSummary
	if maxMsgs <= 0 {
		maxMsgs = DefaultMaxMessagesBeforeSummary
	}
	recentMsgs := cfg.MaxRecentMessages
	if recentMsgs <= 0 {
		recentMsgs = DefaultMaxRecentMessages
	}

	tpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage(SummaryPrompt),
	)

	chain, err := compose.NewChain[map[string]any, *schema.Message]().
		AppendChatTemplate(tpl).
		AppendChatModel(cfg.Model).
		Compile(ctx, compose.WithGraphName("Summarizer"))
	if err != nil {
		return nil, fmt.Errorf("compile summarizer failed: %w", err)
	}

	return &Summarizer{
		chain:      chain,
		maxMsgs:    maxMsgs,
		recentMsgs: recentMsgs,
	}, nil
}

// ShouldSummarize returns true if the messages should be summarized
func (s *Summarizer) ShouldSummarize(msgs []*schema.Message) bool {
	// Count non-system messages
	count := 0
	for _, m := range msgs {
		if m != nil && m.Role != schema.System {
			count++
		}
	}
	return count > s.maxMsgs
}

// Summarize compresses the conversation history
func (s *Summarizer) Summarize(ctx context.Context, msgs []*schema.Message) ([]*schema.Message, error) {
	if len(msgs) == 0 {
		return msgs, nil
	}

	// Separate system messages from conversation
	var systemMsgs []*schema.Message
	var convMsgs []*schema.Message
	for _, m := range msgs {
		if m == nil {
			continue
		}
		if m.Role == schema.System {
			systemMsgs = append(systemMsgs, m)
		} else {
			convMsgs = append(convMsgs, m)
		}
	}

	// If not enough messages, return as-is
	if len(convMsgs) <= s.recentMsgs {
		return msgs, nil
	}

	// Split into older (to summarize) and recent (to keep)
	splitIdx := len(convMsgs) - s.recentMsgs
	olderMsgs := convMsgs[:splitIdx]
	recentMsgs := convMsgs[splitIdx:]

	// Format older messages for summarization
	var sb strings.Builder
	for _, m := range olderMsgs {
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", m.Role, m.Content))
	}

	// Generate summary
	summaryMsg, err := s.chain.Invoke(ctx, map[string]any{
		"conversation": sb.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("summarization failed: %w", err)
	}

	// Build new message list: system + summary + recent
	result := make([]*schema.Message, 0, len(systemMsgs)+1+len(recentMsgs))
	result = append(result, systemMsgs...)

	// Add summary as a special assistant message
	summaryContent := fmt.Sprintf("[对话摘要]\n%s", summaryMsg.Content)
	result = append(result, &schema.Message{
		Role:    schema.Assistant,
		Content: summaryContent,
	})

	result = append(result, recentMsgs...)

	return result, nil
}

// SummarizeIfNeeded checks and summarizes if threshold is exceeded
func (s *Summarizer) SummarizeIfNeeded(ctx context.Context, msgs []*schema.Message) ([]*schema.Message, bool, error) {
	if !s.ShouldSummarize(msgs) {
		return msgs, false, nil
	}

	summarized, err := s.Summarize(ctx, msgs)
	if err != nil {
		return msgs, false, err
	}

	return summarized, true, nil
}
