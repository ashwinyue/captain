package supervisor

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"

	"github.com/tgo/captain/aicenter/internal/eino/agent"
	"github.com/tgo/captain/aicenter/internal/eino/llm"
)

const supervisorInstructionTemplate = `You are a supervisor managing the following agents:

%s
INSTRUCTIONS:
- Analyze the user's request and assign tasks to appropriate agents.
- Assign work to one agent at a time, do not call agents in parallel.
- Do not do any work yourself, always delegate to sub-agents.
- After all tasks are completed, summarize the results and exit.
%s`

type SupervisorConfig struct {
	Name                  string
	SupervisorInstruction string
	SupervisorProvider    *llm.ProviderConfig
	Agents                []*agent.AgentConfig
}

type SupervisorBuilder struct {
	agentBuilder *agent.Builder
	llmFactory   *llm.Factory
}

func NewSupervisorBuilder(agentBuilder *agent.Builder, llmFactory *llm.Factory) *SupervisorBuilder {
	return &SupervisorBuilder{
		agentBuilder: agentBuilder,
		llmFactory:   llmFactory,
	}
}

func (b *SupervisorBuilder) Build(ctx context.Context, cfg *SupervisorConfig) (adk.Agent, error) {
	// Build sub-agents
	subAgents := make([]adk.Agent, 0, len(cfg.Agents))
	for _, agentCfg := range cfg.Agents {
		agent, err := b.agentBuilder.Build(ctx, agentCfg)
		if err != nil {
			return nil, fmt.Errorf("build agent %s: %w", agentCfg.Name, err)
		}
		subAgents = append(subAgents, agent)
	}

	// Build supervisor model
	supervisorModel, err := b.llmFactory.CreateToolCalling(ctx, cfg.SupervisorProvider)
	if err != nil {
		return nil, fmt.Errorf("create supervisor model: %w", err)
	}

	// Build supervisor agent
	sv, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        cfg.Name + "_supervisor",
		Description: fmt.Sprintf("Supervisor for team: %s", cfg.Name),
		Instruction: b.buildInstruction(cfg, cfg.Agents),
		Model:       supervisorModel,
		Exit:        &adk.ExitTool{},
	})
	if err != nil {
		return nil, fmt.Errorf("create supervisor agent: %w", err)
	}

	return supervisor.New(ctx, &supervisor.Config{
		Supervisor: sv,
		SubAgents:  subAgents,
	})
}

func (b *SupervisorBuilder) buildInstruction(cfg *SupervisorConfig, agentConfigs []*agent.AgentConfig) string {
	// Build agent list
	var agentList strings.Builder
	for _, agentCfg := range agentConfigs {
		agentList.WriteString(fmt.Sprintf("- **%s**: %s\n", agentCfg.Name, agentCfg.Description))
	}

	// Build additional instructions
	additional := ""
	if cfg.SupervisorInstruction != "" {
		additional = fmt.Sprintf("\nADDITIONAL INSTRUCTIONS:\n%s", cfg.SupervisorInstruction)
	}

	return fmt.Sprintf(supervisorInstructionTemplate, agentList.String(), additional)
}
