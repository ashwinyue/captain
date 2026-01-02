package orchestration

import (
	"fmt"
	"time"
)

// WorkflowType 工作流类型
type WorkflowType string

const (
	WorkflowSingle       WorkflowType = "single"       // 单 Agent 执行
	WorkflowParallel     WorkflowType = "parallel"     // 并行执行多个 Agent
	WorkflowSequential   WorkflowType = "sequential"   // 串行执行多个 Agent
	WorkflowHierarchical WorkflowType = "hierarchical" // 分层执行：先执行一组，其结果作为下一组的输入
	WorkflowPipeline     WorkflowType = "pipeline"     // 管道执行：数据流式经过多个 Agent
)

// SubQuestion 子问题，用于复杂查询分解
type SubQuestion struct {
	ID              string `json:"id"`                // 唯一标识
	Question        string `json:"question"`          // 分解后的子问题
	Intent          string `json:"intent"`            // 子问题的意图
	AssignedAgentID string `json:"assigned_agent_id"` // 分配的 Agent ID
}

// ExecutionPlan 执行计划
type ExecutionPlan struct {
	Dependencies   [][]string `json:"dependencies"`    // 依赖关系 [[step1], [step2, step3], ...]
	ParallelGroups [][]string `json:"parallel_groups"` // 可并行执行的组
	EstimatedTime  float64    `json:"estimated_time"`  // 预估执行时间（秒）
}

// QueryAnalysisResult 查询分析结果
type QueryAnalysisResult struct {
	// 选中的 Agent IDs
	SelectedAgentIDs []string `json:"selected_agent_ids"`

	// Agent 选择理由
	SelectionReasoning string `json:"selection_reasoning"`

	// 工作流类型
	Workflow WorkflowType `json:"workflow"`

	// 工作流选择理由
	WorkflowReasoning string `json:"workflow_reasoning"`

	// 置信度分数 (0-1)
	ConfidenceScore float64 `json:"confidence_score"`

	// 是否为复杂查询（多意图）
	IsComplex bool `json:"is_complex"`

	// 子问题列表（复杂查询时使用）
	SubQuestions []SubQuestion `json:"sub_questions,omitempty"`

	// 执行计划
	ExecutionPlan *ExecutionPlan `json:"execution_plan,omitempty"`
}

// AgentProfile Agent 简介，用于查询分析
type AgentProfile struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Capabilities []string `json:"capabilities,omitempty"` // Agent 能力标签
}

// AnalysisContext 分析上下文
type AnalysisContext struct {
	ProjectID       string         `json:"project_id"`
	UserQuery       string         `json:"user_query"`
	AvailableAgents []AgentProfile `json:"available_agents"`
	SessionID       string         `json:"session_id,omitempty"`
	UserID          string         `json:"user_id,omitempty"`
}

// AgentResult 单个 Agent 执行结果
type AgentResult struct {
	AgentID     string       `json:"agent_id"`
	Content     string       `json:"content"`
	Error       error        `json:"error,omitempty"`
	SubQuestion *SubQuestion `json:"sub_question,omitempty"`
}

// ExecutionResult 整体执行结果
type ExecutionResult struct {
	Results   []AgentResult `json:"results"`
	Workflow  WorkflowType  `json:"workflow"`
	IsSuccess bool          `json:"is_success"`
}

// GenerateRunID 生成执行 ID
func GenerateRunID() string {
	return "run-" + generateShortID()
}

func generateShortID() string {
	// 简单的 ID 生成
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
