package llm

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

type ProviderKind string

const (
	ProviderOpenAI     ProviderKind = "openai"
	ProviderArk        ProviderKind = "ark"
	ProviderCompatible ProviderKind = "openai_compatible"
	ProviderAnthropic  ProviderKind = "anthropic"
	ProviderGoogle     ProviderKind = "google"
	ProviderDashscope  ProviderKind = "dashscope"
)

type ProviderConfig struct {
	Kind    ProviderKind
	APIKey  string
	Model   string
	BaseURL string
}

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

// Create returns a ToolCallingChatModel (ChatModel is deprecated)
func (f *Factory) Create(ctx context.Context, cfg *ProviderConfig) (model.ToolCallingChatModel, error) {
	return f.CreateToolCalling(ctx, cfg)
}

func (f *Factory) CreateToolCalling(ctx context.Context, cfg *ProviderConfig) (model.ToolCallingChatModel, error) {
	if cfg == nil {
		return nil, fmt.Errorf("provider config is nil")
	}
	switch cfg.Kind {
	case ProviderOpenAI, ProviderCompatible, ProviderDashscope:
		return openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:  cfg.APIKey,
			Model:   cfg.Model,
			BaseURL: cfg.BaseURL,
		})
	case ProviderArk:
		return ark.NewChatModel(ctx, &ark.ChatModelConfig{
			APIKey: cfg.APIKey,
			Model:  cfg.Model,
		})
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Kind)
	}
}

// CreateChatModel returns model.ChatModel for use with react.Agent
func (f *Factory) CreateChatModel(ctx context.Context, cfg *ProviderConfig) (model.ChatModel, error) {
	if cfg == nil {
		return nil, fmt.Errorf("provider config is nil")
	}
	switch cfg.Kind {
	case ProviderOpenAI, ProviderCompatible, ProviderDashscope:
		return openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:  cfg.APIKey,
			Model:   cfg.Model,
			BaseURL: cfg.BaseURL,
		})
	case ProviderArk:
		return ark.NewChatModel(ctx, &ark.ChatModelConfig{
			APIKey: cfg.APIKey,
			Model:  cfg.Model,
		})
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Kind)
	}
}
