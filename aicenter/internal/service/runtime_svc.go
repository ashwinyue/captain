package service

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/adk"
	einoSupervisor "github.com/cloudwego/eino/adk/prebuilt/supervisor"
	einoTool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tgo/captain/aicenter/internal/eino/agent"
	"github.com/tgo/captain/aicenter/internal/eino/llm"
	"github.com/tgo/captain/aicenter/internal/eino/memory"
	"github.com/tgo/captain/aicenter/internal/eino/orchestration"
	"github.com/tgo/captain/aicenter/internal/eino/supervisor"
	"github.com/tgo/captain/aicenter/internal/eino/tool"
	"github.com/tgo/captain/aicenter/internal/model"
	"github.com/tgo/captain/aicenter/internal/pkg/apiserver"
	"github.com/tgo/captain/aicenter/internal/repository"
)

type RuntimeService struct {
	db              *gorm.DB
	teamRepo        *repository.TeamRepository
	aiConfigRepo    *repository.ProjectAIConfigRepository
	providerRepo    *repository.ProviderRepository
	llmFactory      *llm.Factory
	agentBuilder    *agent.Builder
	runner          *supervisor.Runner
	apiserverClient *apiserver.Client  // Client for calling apiserver internal API
	ragURL          string             // RAG service URL for knowledge base tools
	mcpURL          string             // MCP service URL for MCP tools
	redisStore      *memory.RedisStore // Redis store for memory caching
	summarizer      *memory.Summarizer // Conversation summarizer
}

func NewRuntimeService(db *gorm.DB, teamRepo *repository.TeamRepository, aiConfigRepo *repository.ProjectAIConfigRepository, providerRepo *repository.ProviderRepository, ragURL, mcpURL string) *RuntimeService {
	llmFactory := llm.NewFactory()
	agentBuilder := agent.NewBuilder(llmFactory)
	supervisorBuilder := supervisor.NewSupervisorBuilder(agentBuilder, llmFactory)
	runner := supervisor.NewRunner(supervisorBuilder)

	return &RuntimeService{
		db:           db,
		teamRepo:     teamRepo,
		aiConfigRepo: aiConfigRepo,
		providerRepo: providerRepo,
		llmFactory:   llmFactory,
		agentBuilder: agentBuilder,
		runner:       runner,
		ragURL:       ragURL,
		mcpURL:       mcpURL,
	}
}

// SetApiserverClient sets the apiserver client for internal API calls
func (s *RuntimeService) SetApiserverClient(client *apiserver.Client) {
	s.apiserverClient = client
}

// SetRedisStore sets the Redis store for memory caching
func (s *RuntimeService) SetRedisStore(store *memory.RedisStore) {
	s.redisStore = store
}

// SetSummarizer sets the conversation summarizer
func (s *RuntimeService) SetSummarizer(summarizer *memory.Summarizer) {
	s.summarizer = summarizer
}

// SendManualServiceRequest sends a manual service request to the apiserver
func (s *RuntimeService) SendManualServiceRequest(ctx context.Context, visitorID uuid.UUID, reason string) error {
	if s.apiserverClient == nil {
		return nil // Silent fail if not configured
	}
	_, err := s.apiserverClient.SendManualServiceRequest(ctx, visitorID, reason)
	return err
}

// GetMemoryManager returns a memory manager for the given project
// Uses HybridStore (Redis + PostgreSQL) if Redis is available, otherwise PostgresStore
func (s *RuntimeService) GetMemoryManager(projectID uuid.UUID, enablePersistence bool) *memory.Manager {
	return memory.NewManager(&memory.ManagerConfig{
		WindowSize:        10,
		EnablePersistence: enablePersistence,
		DB:                s.db,
		ProjectID:         projectID,
		RedisStore:        s.redisStore, // Will use HybridStore if not nil
		Summarizer:        s.summarizer, // Will enable auto-summarization if not nil
	})
}

// RunRequest matches tgo-ai Python API format
type RunRequest struct {
	TeamID        *string    `json:"team_id"`
	AgentID       *string    `json:"agent_id"`
	AgentIDs      []string   `json:"agent_ids"`
	Message       string     `json:"message"`
	SessionID     *string    `json:"session_id"`
	Stream        bool       `json:"stream"`
	MCPURL        *string    `json:"mcp_url"`
	RAGURL        *string    `json:"rag_url"`
	CollectionIDs []string   `json:"collection_ids"`
	EnableMemory  bool       `json:"enable_memory"`
	VisitorID     *uuid.UUID `json:"visitor_id,omitempty"` // For transfer to human tool
}

type RunResponse struct {
	Content string `json:"content"`
	RunID   string `json:"run_id"`
}

func (s *RuntimeService) Run(ctx context.Context, projectID uuid.UUID, req *RunRequest) (*RunResponse, error) {
	// Get team
	var team *model.Team
	var err error

	if req.TeamID != nil {
		teamID, parseErr := uuid.Parse(*req.TeamID)
		if parseErr != nil {
			return nil, parseErr
		}
		team, err = s.teamRepo.GetWithAgents(ctx, projectID, teamID)
	} else {
		team, err = s.teamRepo.GetDefault(ctx, projectID)
	}
	if err != nil {
		return nil, err
	}

	// Setup memory if enabled
	var memMgr *memory.Manager
	var history []*schema.Message
	sessionID := uuid.New().String()
	if req.SessionID != nil && *req.SessionID != "" {
		sessionID = *req.SessionID
	}
	if req.EnableMemory {
		memMgr = s.GetMemoryManager(projectID, true)
		// Get history before adding new message
		history, _ = memMgr.GetWindowedHistory(ctx, sessionID)
		// Store user message
		_ = memMgr.AddUserMessage(ctx, sessionID, req.Message)
	}

	// Build team config - use service URLs, allow request to override
	mcpURL := s.mcpURL
	ragURL := s.ragURL
	if req.MCPURL != nil && *req.MCPURL != "" {
		mcpURL = *req.MCPURL
	}
	if req.RAGURL != nil && *req.RAGURL != "" {
		ragURL = *req.RAGURL
	}
	teamCfg := s.buildTeamConfigWithVisitor(ctx, projectID, team, mcpURL, ragURL, req.VisitorID)

	// Run with history
	result, err := s.runner.RunWithHistory(ctx, teamCfg, req.Message, history)
	if err != nil {
		return nil, err
	}

	// Store assistant response if memory enabled
	if memMgr != nil && result.Content != "" {
		_ = memMgr.AddAssistantMessage(ctx, sessionID, result.Content)
	}

	return &RunResponse{
		Content: result.Content,
		RunID:   uuid.New().String(),
	}, nil
}

// RunWithReactAgent runs a single agent with tools using ReAct pattern
// This is recommended for agents with RAG/tool support
func (s *RuntimeService) RunWithReactAgent(ctx context.Context, projectID uuid.UUID, agentID string, message string, instruction string, tools []einoTool.BaseTool) (*RunResponse, error) {
	// Get project default provider config
	var providerCfg *llm.ProviderConfig
	if aiConfig, err := s.aiConfigRepo.GetByProjectID(ctx, projectID); err == nil && aiConfig != nil {
		if aiConfig.DefaultChatProviderID != nil {
			if provider, err := s.providerRepo.GetByID(ctx, projectID, *aiConfig.DefaultChatProviderID); err == nil && provider != nil {
				providerCfg = &llm.ProviderConfig{
					Kind:    llm.ProviderKind(provider.ProviderKind),
					APIKey:  provider.APIKey,
					Model:   aiConfig.DefaultChatModel,
					BaseURL: provider.APIBaseURL,
				}
			}
		}
	}

	if providerCfg == nil {
		return nil, fmt.Errorf("no default provider configured for project")
	}

	// Build ReAct agent config
	agentCfg := &agent.AgentConfig{
		Name:        "assistant",
		Description: "AI assistant with knowledge base tools",
		Instruction: instruction,
		Provider:    providerCfg,
		Tools:       tools,
	}

	// Create ReAct agent
	reactAgent, err := s.agentBuilder.BuildReactAgent(ctx, agentCfg)
	if err != nil {
		return nil, fmt.Errorf("build react agent: %w", err)
	}

	// Build messages with system instruction (required for react agent to use tools properly)
	var messages []*schema.Message

	// Add system instruction that tells the agent to use tools
	systemPrompt := instruction
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant with access to knowledge base search tools. When answering questions, always search the knowledge base first to find relevant information."
	} else {
		systemPrompt += "\n\nIMPORTANT: You have access to knowledge base search tools. When answering questions, ALWAYS use the search tools to find relevant information before responding."
	}
	messages = append(messages, schema.SystemMessage(systemPrompt))
	messages = append(messages, schema.UserMessage(message))

	log.Printf("[DEBUG] Running ReAct agent with %d messages", len(messages))

	// Run agent
	resp, err := reactAgent.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("react agent generate: %w", err)
	}

	return &RunResponse{
		Content: resp.Content,
		RunID:   uuid.New().String(),
	}, nil
}

// RunWithReactAgentAndMemory runs ReAct agent with session memory support
func (s *RuntimeService) RunWithReactAgentAndMemory(ctx context.Context, projectID uuid.UUID, agentID string, message string, instruction string, tools []einoTool.BaseTool, sessionID string, enableMemory bool) (*RunResponse, error) {
	// Get provider config
	providerCfg, err := s.getDefaultProviderConfig(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get provider config: %w", err)
	}

	// Build ReAct agent config
	agentCfg := &agent.AgentConfig{
		Name:        "assistant",
		Description: "AI assistant with knowledge base tools",
		Instruction: instruction,
		Provider:    providerCfg,
		Tools:       tools,
	}

	// Create ReAct agent
	reactAgent, err := s.agentBuilder.BuildReactAgent(ctx, agentCfg)
	if err != nil {
		return nil, fmt.Errorf("build react agent: %w", err)
	}

	// Build messages with system instruction
	var messages []*schema.Message
	systemPrompt := instruction
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant with access to knowledge base search tools. When answering questions, always search the knowledge base first to find relevant information."
	} else {
		systemPrompt += "\n\nIMPORTANT: You have access to knowledge base search tools. When answering questions, ALWAYS use the search tools to find relevant information before responding."
	}
	messages = append(messages, schema.SystemMessage(systemPrompt))

	// Add conversation history if memory is enabled
	if enableMemory && sessionID != "" {
		memMgr := s.GetMemoryManager(projectID, true)
		history, err := memMgr.GetWindowedHistory(ctx, sessionID)
		if err != nil {
			log.Printf("[WARN] Failed to get history: %v", err)
		} else if len(history) > 0 {
			log.Printf("[DEBUG] Loaded %d messages from session %s", len(history), sessionID)
			messages = append(messages, history...)
		}
	}

	// Add current user message
	messages = append(messages, schema.UserMessage(message))

	log.Printf("[DEBUG] Running ReAct agent with %d messages (memory=%v, session=%s)", len(messages), enableMemory, sessionID)

	// Run agent
	resp, err := reactAgent.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("react agent generate: %w", err)
	}

	// Save to memory if enabled
	if enableMemory && sessionID != "" {
		memMgr := s.GetMemoryManager(projectID, true)
		if err := memMgr.AddUserMessage(ctx, sessionID, message); err != nil {
			log.Printf("[WARN] Failed to save user message: %v", err)
		}
		if err := memMgr.AddAssistantMessage(ctx, sessionID, resp.Content); err != nil {
			log.Printf("[WARN] Failed to save assistant message: %v", err)
		}
	}

	return &RunResponse{
		Content: resp.Content,
		RunID:   uuid.New().String(),
	}, nil
}

// RunWithAgentTools loads agent's tools and runs using ReAct pattern
func (s *RuntimeService) RunWithAgentTools(ctx context.Context, projectID uuid.UUID, agentID string, message string, sessionID string, enableMemory bool) (*RunResponse, error) {
	// Get agent with collections
	agentUUID, err := uuid.Parse(agentID)
	if err != nil {
		return nil, fmt.Errorf("invalid agent_id: %w", err)
	}

	// Query agent and its collections from database
	var dbAgent model.Agent
	if err := s.db.WithContext(ctx).Where("id = ?", agentUUID).First(&dbAgent).Error; err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// Load agent's collections
	var collections []model.AgentCollection
	s.db.WithContext(ctx).Where("agent_id = ? AND is_enabled = ?", agentUUID, true).Find(&collections)

	// Load RAG tools from collections
	var collectionIDs []string
	for _, c := range collections {
		collectionIDs = append(collectionIDs, c.CollectionID)
	}

	log.Printf("[DEBUG] Agent %s has %d enabled collections, ragURL=%s", dbAgent.Name, len(collectionIDs), s.ragURL)

	tools, err := tool.LoadRAGTools(ctx, s.ragURL, collectionIDs)
	if err != nil {
		log.Printf("[WARN] Failed to load RAG tools: %v", err)
	} else {
		log.Printf("[DEBUG] Loaded %d RAG tools", len(tools))
	}

	// If no tools, fall back to regular run
	if len(tools) == 0 {
		log.Printf("[DEBUG] No tools loaded, falling back to regular run")
		return s.Run(ctx, projectID, &RunRequest{Message: message, SessionID: &sessionID, EnableMemory: enableMemory})
	}

	// Use RunWithReactAgent with memory support
	return s.RunWithReactAgentAndMemory(ctx, projectID, agentID, message, dbAgent.Instruction, tools, sessionID, enableMemory)
}

// RunWithQueryAnalyzer 使用 QueryAnalyzer 智能路由查询
// 首先分析查询意图和复杂度，然后选择合适的执行策略
func (s *RuntimeService) RunWithQueryAnalyzer(ctx context.Context, projectID uuid.UUID, message string) (*RunResponse, error) {
	log.Printf("[RunWithQueryAnalyzer] Starting analysis for: %s", message)

	// 1. 获取项目的默认 provider 配置
	providerCfg, err := s.getDefaultProviderConfig(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get provider config: %w", err)
	}

	// 2. 创建用于分析的 ChatModel
	chatModel, err := s.llmFactory.CreateChatModel(ctx, providerCfg)
	if err != nil {
		return nil, fmt.Errorf("create chat model: %w", err)
	}

	// 3. 获取可用的 Agents
	team, err := s.teamRepo.GetDefault(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get default team: %w", err)
	}

	// 构建 Agent 简介列表
	agentProfiles := make([]orchestration.AgentProfile, 0, len(team.Agents))
	for _, a := range team.Agents {
		agentProfiles = append(agentProfiles, orchestration.AgentProfile{
			ID:          a.ID.String(),
			Name:        a.Name,
			Description: a.Description,
		})
	}

	// 4. 创建 QueryAnalyzer 并分析
	analyzer := orchestration.NewQueryAnalyzer(chatModel)
	analysisCtx := &orchestration.AnalysisContext{
		ProjectID:       projectID.String(),
		UserQuery:       message,
		AvailableAgents: agentProfiles,
	}

	result, err := analyzer.Analyze(ctx, analysisCtx)
	if err != nil {
		log.Printf("[QueryAnalyzer] Analysis failed: %v, falling back to default", err)
		// 降级到默认处理
		return s.Run(ctx, projectID, &RunRequest{Message: message})
	}

	log.Printf("[QueryAnalyzer] Result: workflow=%s, agents=%v, is_complex=%v, confidence=%.2f",
		result.Workflow, result.SelectedAgentIDs, result.IsComplex, result.ConfidenceScore)

	// 5. 根据分析结果执行
	switch result.Workflow {
	case orchestration.WorkflowSingle:
		// 单 Agent 执行
		if len(result.SelectedAgentIDs) > 0 {
			return s.RunWithAgentTools(ctx, projectID, result.SelectedAgentIDs[0], message, "", false)
		}
		return s.Run(ctx, projectID, &RunRequest{Message: message})

	case orchestration.WorkflowParallel, orchestration.WorkflowSequential:
		// 并行/串行多 Agent 执行
		return s.executeMultiAgent(ctx, projectID, result, message)

	default:
		return s.Run(ctx, projectID, &RunRequest{Message: message})
	}
}

// executeMultiAgent 执行多 Agent 工作流（按 eino-examples 最佳实践）
func (s *RuntimeService) executeMultiAgent(ctx context.Context, projectID uuid.UUID, analysis *orchestration.QueryAnalysisResult, message string) (*RunResponse, error) {
	log.Printf("[MultiAgent] Starting %s execution with %d agents (eino ADK)", analysis.Workflow, len(analysis.SelectedAgentIDs))

	// 根据工作流类型选择执行方式（全部使用 eino ADK）
	switch analysis.Workflow {
	case orchestration.WorkflowParallel:
		return s.executeParallelWithEino(ctx, projectID, analysis, message)
	case orchestration.WorkflowSequential, orchestration.WorkflowHierarchical, orchestration.WorkflowPipeline:
		return s.executeSequentialWithEino(ctx, projectID, analysis, message)
	default:
		// 默认使用并行执行
		return s.executeParallelWithEino(ctx, projectID, analysis, message)
	}
}

// executeParallelWithEino 使用 eino NewParallelAgent 并行执行
func (s *RuntimeService) executeParallelWithEino(ctx context.Context, projectID uuid.UUID, analysis *orchestration.QueryAnalysisResult, message string) (*RunResponse, error) {
	log.Printf("[MultiAgent] Parallel execution with eino NewParallelAgent")

	// 1. 构建所有子 Agent
	subAgents, err := s.buildSubAgents(ctx, projectID, analysis.SelectedAgentIDs)
	if err != nil {
		return nil, fmt.Errorf("build sub agents: %w", err)
	}

	// 2. 使用 eino NewParallelAgent
	parallelAgent, err := adk.NewParallelAgent(ctx, &adk.ParallelAgentConfig{
		Name:        "ParallelExecutor",
		Description: "Parallel execution of multiple agents",
		SubAgents:   subAgents,
	})
	if err != nil {
		return nil, fmt.Errorf("create parallel agent: %w", err)
	}

	// 3. 使用 adk.Runner 执行
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		EnableStreaming: false,
		Agent:           parallelAgent,
	})

	iter := runner.Query(ctx, message)
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

	log.Printf("[MultiAgent] Parallel execution completed, result length: %d", len(lastMsg.Content))

	return &RunResponse{
		Content: lastMsg.Content,
		RunID:   orchestration.GenerateRunID(),
	}, nil
}

// executeSequentialWithEino 使用 eino supervisor 串行执行
func (s *RuntimeService) executeSequentialWithEino(ctx context.Context, projectID uuid.UUID, analysis *orchestration.QueryAnalysisResult, message string) (*RunResponse, error) {
	log.Printf("[MultiAgent] Sequential execution with eino Supervisor")

	// 1. 构建所有子 Agent
	subAgents, err := s.buildSubAgents(ctx, projectID, analysis.SelectedAgentIDs)
	if err != nil {
		return nil, fmt.Errorf("build sub agents: %w", err)
	}

	// 2. 获取 Supervisor 模型
	providerCfg, err := s.getDefaultProviderConfig(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get provider config: %w", err)
	}

	supervisorModel, err := s.llmFactory.CreateToolCalling(ctx, providerCfg)
	if err != nil {
		return nil, fmt.Errorf("create supervisor model: %w", err)
	}

	// 3. 构建 Supervisor Agent（按 eino-examples 模式）
	supervisorAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "SequentialCoordinator",
		Description: "Coordinates sequential execution of agents",
		Instruction: s.buildSequentialInstruction(subAgents),
		Model:       supervisorModel,
		Exit:        &adk.ExitTool{},
	})
	if err != nil {
		return nil, fmt.Errorf("create supervisor agent: %w", err)
	}

	// 4. 使用 supervisor.New
	coordinatedAgent, err := einoSupervisor.New(ctx, &einoSupervisor.Config{
		Supervisor: supervisorAgent,
		SubAgents:  subAgents,
	})
	if err != nil {
		return nil, fmt.Errorf("create coordinated agent: %w", err)
	}

	// 5. 使用 adk.Runner 执行
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		EnableStreaming: false,
		Agent:           coordinatedAgent,
	})

	iter := runner.Query(ctx, message)
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

	log.Printf("[MultiAgent] Sequential execution completed, result length: %d", len(lastMsg.Content))

	return &RunResponse{
		Content: lastMsg.Content,
		RunID:   orchestration.GenerateRunID(),
	}, nil
}

// buildSubAgents 构建子 Agent 列表
func (s *RuntimeService) buildSubAgents(ctx context.Context, projectID uuid.UUID, agentIDs []string) ([]adk.Agent, error) {
	subAgents := make([]adk.Agent, 0, len(agentIDs))

	for _, agentIDStr := range agentIDs {
		agentID, err := uuid.Parse(agentIDStr)
		if err != nil {
			log.Printf("[MultiAgent] Invalid agent ID: %s, skipping", agentIDStr)
			continue
		}

		// 获取 Agent 配置
		agentCfg, err := s.buildAgentConfig(ctx, projectID, agentID)
		if err != nil {
			log.Printf("[MultiAgent] Failed to build agent config for %s: %v, skipping", agentIDStr, err)
			continue
		}

		// 使用 agentBuilder 构建 Agent（遵循 eino-examples 模式）
		agent, err := s.agentBuilder.Build(ctx, agentCfg)
		if err != nil {
			log.Printf("[MultiAgent] Failed to build agent %s: %v, skipping", agentIDStr, err)
			continue
		}

		subAgents = append(subAgents, agent)
	}

	if len(subAgents) == 0 {
		return nil, fmt.Errorf("no valid agents built")
	}

	return subAgents, nil
}

// buildSequentialInstruction 构建串行执行指令
func (s *RuntimeService) buildSequentialInstruction(agents []adk.Agent) string {
	instruction := `You are a coordinator managing sequential task execution.

INSTRUCTIONS:
1. Execute agents ONE BY ONE in the order they are available
2. Pass the output of each agent to the next agent as context
3. Do NOT call agents in parallel
4. After all agents complete, summarize the results and exit
5. Do not do any work yourself, always delegate to sub-agents`

	return instruction
}

// buildAgentConfig 构建 Agent 配置（用于 eino ADK）
func (s *RuntimeService) buildAgentConfig(ctx context.Context, projectID uuid.UUID, agentID uuid.UUID) (*agent.AgentConfig, error) {
	// 1. 查询 Agent
	var dbAgent model.Agent
	if err := s.db.WithContext(ctx).Where("id = ?", agentID).First(&dbAgent).Error; err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	// 2. 获取 Provider 配置
	providerCfg, err := s.getDefaultProviderConfig(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get provider config: %w", err)
	}

	// 3. 加载 RAG 工具
	var collections []model.AgentCollection
	s.db.WithContext(ctx).Where("agent_id = ? AND is_enabled = ?", agentID, true).Find(&collections)

	var collectionIDs []string
	for _, c := range collections {
		collectionIDs = append(collectionIDs, c.CollectionID)
	}

	var tools []einoTool.BaseTool
	if len(collectionIDs) > 0 {
		ragTools, err := tool.LoadRAGTools(ctx, s.ragURL, collectionIDs)
		if err != nil {
			log.Printf("[buildAgentConfig] Failed to load RAG tools: %v", err)
		} else {
			tools = append(tools, ragTools...)
		}
	}

	return &agent.AgentConfig{
		Name:        dbAgent.Name,
		Description: dbAgent.Description,
		Instruction: dbAgent.Instruction,
		Provider:    providerCfg,
		Tools:       tools,
	}, nil
}

// getDefaultProviderConfig 获取项目的默认 provider 配置
func (s *RuntimeService) getDefaultProviderConfig(ctx context.Context, projectID uuid.UUID) (*llm.ProviderConfig, error) {
	aiConfig, err := s.aiConfigRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get AI config: %w", err)
	}
	if aiConfig == nil || aiConfig.DefaultChatProviderID == nil {
		return nil, fmt.Errorf("no default provider configured")
	}

	provider, err := s.providerRepo.GetByID(ctx, projectID, *aiConfig.DefaultChatProviderID)
	if err != nil {
		return nil, fmt.Errorf("get provider: %w", err)
	}

	return &llm.ProviderConfig{
		Kind:    llm.ProviderKind(provider.ProviderKind),
		APIKey:  provider.APIKey,
		Model:   aiConfig.DefaultChatModel,
		BaseURL: provider.APIBaseURL,
	}, nil
}

func (s *RuntimeService) Stream(ctx context.Context, projectID uuid.UUID, req *RunRequest, callback supervisor.StreamCallback) error {
	// Get team
	var team *model.Team
	var err error

	if req.TeamID != nil {
		teamID, parseErr := uuid.Parse(*req.TeamID)
		if parseErr != nil {
			return parseErr
		}
		team, err = s.teamRepo.GetWithAgents(ctx, projectID, teamID)
	} else {
		team, err = s.teamRepo.GetDefault(ctx, projectID)
	}
	if err != nil {
		return err
	}

	// Setup memory if enabled
	var memMgr *memory.Manager
	sessionID := uuid.New().String()
	if req.SessionID != nil && *req.SessionID != "" {
		sessionID = *req.SessionID
	}
	if req.EnableMemory {
		memMgr = s.GetMemoryManager(projectID, true)
		// Store user message
		_ = memMgr.AddUserMessage(ctx, sessionID, req.Message)
	}

	// Build team config - use service URLs, allow request to override
	mcpURL := s.mcpURL
	ragURL := s.ragURL
	if req.MCPURL != nil && *req.MCPURL != "" {
		mcpURL = *req.MCPURL
	}
	if req.RAGURL != nil && *req.RAGURL != "" {
		ragURL = *req.RAGURL
	}
	teamCfg := s.buildTeamConfigWithVisitor(ctx, projectID, team, mcpURL, ragURL, req.VisitorID)

	// Wrap callback to capture final response for memory
	var finalContent string
	wrappedCallback := func(event *adk.AgentEvent) error {
		// Capture content from message output
		if event.Output != nil && event.Output.MessageOutput != nil {
			if msg := event.Output.MessageOutput.Message; msg != nil {
				finalContent = msg.Content
			}
		}
		return callback(event)
	}

	// Stream
	err = s.runner.Stream(ctx, teamCfg, req.Message, wrappedCallback)

	// Store assistant response if memory enabled
	if memMgr != nil && finalContent != "" {
		_ = memMgr.AddAssistantMessage(ctx, sessionID, finalContent)
	}

	return err
}

func (s *RuntimeService) buildTeamConfig(ctx context.Context, projectID uuid.UUID, team *model.Team, mcpURL, ragURL string) *supervisor.SupervisorConfig {
	return s.buildTeamConfigWithVisitor(ctx, projectID, team, mcpURL, ragURL, nil)
}

func (s *RuntimeService) buildTeamConfigWithVisitor(ctx context.Context, projectID uuid.UUID, team *model.Team, mcpURL, ragURL string, visitorID *uuid.UUID) *supervisor.SupervisorConfig {
	// Get project default provider config
	var defaultProviderCfg *llm.ProviderConfig
	if aiConfig, err := s.aiConfigRepo.GetByProjectID(ctx, projectID); err == nil && aiConfig != nil {
		if aiConfig.DefaultChatProviderID != nil {
			if provider, err := s.providerRepo.GetByID(ctx, projectID, *aiConfig.DefaultChatProviderID); err == nil && provider != nil {
				defaultProviderCfg = &llm.ProviderConfig{
					Kind:    llm.ProviderKind(provider.ProviderKind),
					APIKey:  provider.APIKey,
					Model:   aiConfig.DefaultChatModel,
					BaseURL: provider.APIBaseURL,
				}
			}
		}
	}

	// Build agent configs
	agentConfigs := make([]*agent.AgentConfig, 0, len(team.Agents))
	for _, a := range team.Agents {
		if !a.IsEnabled {
			continue
		}

		var providerCfg *llm.ProviderConfig
		if a.LLMProvider != nil {
			providerCfg = &llm.ProviderConfig{
				Kind:    llm.ProviderKind(a.LLMProvider.ProviderKind),
				APIKey:  a.LLMProvider.APIKey,
				Model:   a.LLMProvider.DefaultModel,
				BaseURL: a.LLMProvider.APIBaseURL,
			}
		} else if defaultProviderCfg != nil {
			// Use project default provider
			providerCfg = defaultProviderCfg
		}

		// Load RAG tools from agent's collections
		var collectionIDs []string
		for _, c := range a.Collections {
			if c.IsEnabled {
				collectionIDs = append(collectionIDs, c.CollectionID)
			}
		}

		log.Printf("[DEBUG] Agent %s has %d collections, ragURL=%s", a.Name, len(collectionIDs), ragURL)
		tools, err := tool.LoadRAGTools(ctx, ragURL, collectionIDs)
		if err != nil {
			log.Printf("[ERROR] Failed to load RAG tools for agent %s: %v", a.Name, err)
		} else {
			log.Printf("[DEBUG] Loaded %d RAG tools for agent %s", len(tools), a.Name)
		}
		mcpTools, _ := tool.LoadMCPTools(ctx, mcpURL)
		tools = append(tools, mcpTools...)

		// Add transfer_to_human tool if visitor context is available
		instruction := a.Instruction
		if visitorID != nil && s.apiserverClient != nil {
			vid := *visitorID
			transferTool := tool.NewTransferHumanTool(func(ctx context.Context, reason string) error {
				return s.SendManualServiceRequest(ctx, vid, reason)
			})
			tools = append(tools, transferTool)
			// Append transfer tool usage instruction
			instruction += `

IMPORTANT: You have access to the transfer_to_human tool. When the user explicitly requests human assistance (e.g., "转人工", "人工客服", "human agent", "speak to agent"), you MUST call the transfer_to_human tool immediately with the reason. Do NOT ask for more details - just transfer them.`
		}

		agentConfigs = append(agentConfigs, &agent.AgentConfig{
			Name:        a.Name,
			Description: a.Description,
			Instruction: instruction,
			Provider:    providerCfg,
			Tools:       tools,
		})
	}

	// Create default agent if no agents are configured
	if len(agentConfigs) == 0 && defaultProviderCfg != nil {
		var defaultTools []einoTool.BaseTool
		// Add transfer_to_human tool if visitor context is available
		if visitorID != nil && s.apiserverClient != nil {
			vid := *visitorID
			transferTool := tool.NewTransferHumanTool(func(ctx context.Context, reason string) error {
				return s.SendManualServiceRequest(ctx, vid, reason)
			})
			defaultTools = append(defaultTools, transferTool)
		}

		agentConfigs = append(agentConfigs, &agent.AgentConfig{
			Name:        "Assistant",
			Description: "A helpful AI assistant that can answer questions and help users. Can transfer to human agents when needed.",
			Instruction: `You are a helpful customer service assistant. Follow these rules:
1. Be polite and helpful to users
2. Answer questions to the best of your ability
3. When the user explicitly requests human assistance (e.g., "转人工", "人工客服", "human agent"), use the transfer_to_human tool immediately
4. Do not ask for more details when the user clearly wants human assistance - just transfer them`,
			Provider: defaultProviderCfg,
			Tools:    defaultTools,
		})
	}

	// Build supervisor provider config
	var supervisorProvider *llm.ProviderConfig
	if team.SupervisorLLM != nil {
		supervisorProvider = &llm.ProviderConfig{
			Kind:    llm.ProviderKind(team.SupervisorLLM.ProviderKind),
			APIKey:  team.SupervisorLLM.APIKey,
			Model:   team.SupervisorLLM.DefaultModel,
			BaseURL: team.SupervisorLLM.APIBaseURL,
		}
	} else if len(agentConfigs) > 0 && agentConfigs[0].Provider != nil {
		// Default to first agent's provider
		supervisorProvider = agentConfigs[0].Provider
	} else if defaultProviderCfg != nil {
		// Use project default provider
		supervisorProvider = defaultProviderCfg
	}

	return &supervisor.SupervisorConfig{
		Name:                  team.Name,
		SupervisorInstruction: team.SupervisorInstruction,
		SupervisorProvider:    supervisorProvider,
		Agents:                agentConfigs,
	}
}

// StreamEvent converts ADK event to SSE format
type StreamEvent struct {
	Type      string      `json:"type"`
	AgentName string      `json:"agent_name,omitempty"`
	Content   string      `json:"content,omitempty"`
	Error     string      `json:"error,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

func ConvertADKEvent(event *adk.AgentEvent) *StreamEvent {
	se := &StreamEvent{
		AgentName: event.AgentName,
	}

	if event.Err != nil {
		se.Type = "error"
		se.Error = event.Err.Error()
		return se
	}

	if event.Output != nil && event.Output.MessageOutput != nil {
		if msg := event.Output.MessageOutput.Message; msg != nil {
			se.Type = "message"
			se.Content = msg.Content
		}
	}

	if event.Action != nil {
		if event.Action.Exit {
			se.Type = "exit"
		} else if event.Action.TransferToAgent != nil {
			se.Type = "transfer"
			se.Content = event.Action.TransferToAgent.DestAgentName
		}
	}

	return se
}

// GetEventContent extracts content from an ADK event
func GetEventContent(event *adk.AgentEvent) string {
	if event.Output != nil && event.Output.MessageOutput != nil {
		if msg := event.Output.MessageOutput.Message; msg != nil {
			return msg.Content
		}
	}
	return ""
}

// Cancel cancels a running execution (placeholder - context cancellation is preferred)
func (s *RuntimeService) Cancel(ctx context.Context, runID string) (bool, string) {
	// In Go, cancellation is typically handled via context cancellation
	// The runID would be used to lookup and cancel the associated context
	// For now, return false as we don't have a run registry
	return false, "cancellation not implemented - use context cancellation"
}
