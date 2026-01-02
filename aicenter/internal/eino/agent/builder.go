package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"

	"github.com/tgo/captain/aicenter/internal/eino/llm"
)

type AgentConfig struct {
	Name        string
	Description string
	Instruction string
	Provider    *llm.ProviderConfig
	Tools       []tool.BaseTool
}

type Builder struct {
	llmFactory *llm.Factory
}

func NewBuilder(llmFactory *llm.Factory) *Builder {
	return &Builder{llmFactory: llmFactory}
}

func (b *Builder) Build(ctx context.Context, cfg *AgentConfig) (adk.Agent, error) {
	chatModel, err := b.llmFactory.CreateToolCalling(ctx, cfg.Provider)
	if err != nil {
		return nil, fmt.Errorf("create chat model: %w", err)
	}

	// Log tool info for debugging
	if len(cfg.Tools) > 0 {
		log.Printf("[DEBUG] Building agent %s with %d tools", cfg.Name, len(cfg.Tools))
		for _, t := range cfg.Tools {
			info, _ := t.Info(ctx)
			if info != nil {
				log.Printf("[DEBUG] Tool: %s - %s", info.Name, info.Desc)
			}
		}
	}

	toolsConfig := adk.ToolsConfig{
		ToolsNodeConfig: compose.ToolsNodeConfig{
			Tools: cfg.Tools,
			UnknownToolsHandler: func(ctx context.Context, name, input string) (string, error) {
				return fmt.Sprintf("unknown tool: %s", name), nil
			},
		},
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        cfg.Name,
		Description: cfg.Description,
		Instruction: cfg.Instruction,
		Model:       chatModel,
		ToolsConfig: toolsConfig,
	})
}

// BuildReactAgent creates a ReAct agent for direct tool calling (not for supervisor)
// Use this when you need an agent that actively uses tools to answer questions
func (b *Builder) BuildReactAgent(ctx context.Context, cfg *AgentConfig) (*react.Agent, error) {
	chatModel, err := b.llmFactory.CreateChatModel(ctx, cfg.Provider)
	if err != nil {
		return nil, fmt.Errorf("create chat model: %w", err)
	}

	log.Printf("[DEBUG] Building ReAct agent %s with %d tools", cfg.Name, len(cfg.Tools))
	for _, t := range cfg.Tools {
		info, _ := t.Info(ctx)
		if info != nil {
			log.Printf("[DEBUG] Tool: %s - %s", info.Name, info.Desc)
		}
	}

	reactConfig := &react.AgentConfig{
		Model:   chatModel,
		MaxStep: 10,
		ToolsConfig: compose.ToolsNodeConfig{
			Tools: cfg.Tools,
		},
	}

	return react.NewAgent(ctx, reactConfig)
}
