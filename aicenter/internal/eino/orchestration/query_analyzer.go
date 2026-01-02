package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// QueryAnalyzer 查询分析器，负责分析用户查询的意图和复杂度
type QueryAnalyzer struct {
	chatModel    model.ChatModel
	intentRouter *IntentRouter
}

// NewQueryAnalyzer 创建查询分析器
func NewQueryAnalyzer(chatModel model.ChatModel) *QueryAnalyzer {
	return &QueryAnalyzer{
		chatModel:    chatModel,
		intentRouter: NewIntentRouter(),
	}
}

// Analyze 分析用户查询
func (qa *QueryAnalyzer) Analyze(ctx context.Context, analysisCtx *AnalysisContext) (*QueryAnalysisResult, error) {
	// 1. 先尝试关键词快速匹配（参考 eino router 模式）
	quickResult := qa.intentRouter.Match(analysisCtx.UserQuery)
	if quickResult.Matched && quickResult.SkipLLM {
		log.Printf("[QueryAnalyzer] Quick match: intent=%s, reason=%s", quickResult.Intent, quickResult.Reason)
		return qa.buildQuickResult(quickResult, analysisCtx), nil
	}

	if quickResult.Matched {
		log.Printf("[QueryAnalyzer] Partial match: intent=%s, continue with LLM", quickResult.Intent)
	}

	// 2. 未匹配或需要 LLM 进一步分析
	// 构建 Agent 简介文本
	agentProfilesText := qa.buildAgentProfilesText(analysisCtx.AvailableAgents)

	// 构建分析 prompt
	prompt := strings.ReplaceAll(QueryAnalyzerPrompt, "{agent_profiles}", agentProfilesText)
	prompt = strings.ReplaceAll(prompt, "{user_query}", analysisCtx.UserQuery)

	log.Printf("[QueryAnalyzer] Analyzing query: %s", analysisCtx.UserQuery)
	log.Printf("[QueryAnalyzer] Available agents: %d", len(analysisCtx.AvailableAgents))

	// 调用 LLM 进行分析
	messages := []*schema.Message{
		schema.UserMessage(prompt),
	}

	resp, err := qa.chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM generate failed: %w", err)
	}

	log.Printf("[QueryAnalyzer] LLM response: %s", resp.Content)

	// 解析 LLM 响应
	result, err := qa.parseResponse(resp.Content)
	if err != nil {
		log.Printf("[QueryAnalyzer] Parse error: %v, using fallback", err)
		// 解析失败时使用降级策略
		return qa.fallbackResult(analysisCtx), nil
	}

	// 验证结果
	if err := qa.validateResult(result, analysisCtx); err != nil {
		log.Printf("[QueryAnalyzer] Validation error: %v, using fallback", err)
		return qa.fallbackResult(analysisCtx), nil
	}

	log.Printf("[QueryAnalyzer] Analysis result: workflow=%s, agents=%v, is_complex=%v",
		result.Workflow, result.SelectedAgentIDs, result.IsComplex)

	return result, nil
}

// buildAgentProfilesText 构建 Agent 简介文本
func (qa *QueryAnalyzer) buildAgentProfilesText(agents []AgentProfile) string {
	if len(agents) == 0 {
		return "无可用 Agent"
	}

	var sb strings.Builder
	for _, agent := range agents {
		text := AgentProfileTemplate
		text = strings.ReplaceAll(text, "{id}", agent.ID)
		text = strings.ReplaceAll(text, "{name}", agent.Name)
		text = strings.ReplaceAll(text, "{description}", agent.Description)
		sb.WriteString(text)
		sb.WriteString("\n")
	}
	return sb.String()
}

// parseResponse 解析 LLM 响应
func (qa *QueryAnalyzer) parseResponse(content string) (*QueryAnalysisResult, error) {
	// 清理响应内容，提取 JSON
	content = strings.TrimSpace(content)

	// 尝试提取 JSON 块
	if idx := strings.Index(content, "{"); idx >= 0 {
		content = content[idx:]
	}
	if idx := strings.LastIndex(content, "}"); idx >= 0 {
		content = content[:idx+1]
	}

	var result QueryAnalysisResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("JSON unmarshal failed: %w", err)
	}

	return &result, nil
}

// validateResult 验证分析结果
func (qa *QueryAnalyzer) validateResult(result *QueryAnalysisResult, ctx *AnalysisContext) error {
	// 验证 selected_agent_ids 不为空
	if len(result.SelectedAgentIDs) == 0 {
		return fmt.Errorf("no agents selected")
	}

	// 验证选中的 Agent 存在于可用列表中
	availableIDs := make(map[string]bool)
	for _, agent := range ctx.AvailableAgents {
		availableIDs[agent.ID] = true
	}

	for _, id := range result.SelectedAgentIDs {
		if !availableIDs[id] {
			return fmt.Errorf("selected agent %s not in available list", id)
		}
	}

	// 验证 workflow 类型
	switch result.Workflow {
	case WorkflowSingle, WorkflowParallel, WorkflowSequential:
		// OK
	default:
		return fmt.Errorf("invalid workflow type: %s", result.Workflow)
	}

	// 验证 confidence_score 范围
	if result.ConfidenceScore < 0 || result.ConfidenceScore > 1 {
		result.ConfidenceScore = 0.5 // 修正为默认值
	}

	return nil
}

// fallbackResult 降级策略：选择第一个可用的 Agent
func (qa *QueryAnalyzer) fallbackResult(ctx *AnalysisContext) *QueryAnalysisResult {
	result := &QueryAnalysisResult{
		Workflow:           WorkflowSingle,
		WorkflowReasoning:  "Fallback: 使用默认单 Agent 模式",
		ConfidenceScore:    0.5,
		IsComplex:          false,
		SelectionReasoning: "Fallback: 解析失败，选择默认 Agent",
	}

	if len(ctx.AvailableAgents) > 0 {
		result.SelectedAgentIDs = []string{ctx.AvailableAgents[0].ID}
	}

	return result
}

// AnalyzeSimple 简化版分析：仅判断是否需要使用工具
// 用于快速判断是否应该使用 RAG 等工具
func (qa *QueryAnalyzer) AnalyzeSimple(ctx context.Context, query string, agents []AgentProfile) (*QueryAnalysisResult, error) {
	analysisCtx := &AnalysisContext{
		UserQuery:       query,
		AvailableAgents: agents,
	}
	return qa.Analyze(ctx, analysisCtx)
}

// buildQuickResult 根据快速匹配结果构建分析结果
// 用于关键词匹配成功时快速返回，无需调用 LLM
func (qa *QueryAnalyzer) buildQuickResult(quickResult *QuickMatchResult, analysisCtx *AnalysisContext) *QueryAnalysisResult {
	result := &QueryAnalysisResult{
		Workflow:        WorkflowSingle,
		ConfidenceScore: 0.95, // 关键词匹配置信度高
		IsComplex:       false,
	}

	switch quickResult.Intent {
	case IntentHuman:
		// 转人工：不需要 Agent，直接返回特殊标记
		result.WorkflowReasoning = "关键词快速匹配：用户请求转人工"
		result.SelectionReasoning = "用户明确要求转人工服务"
		result.SelectedAgentIDs = []string{} // 空表示转人工
		result.Workflow = WorkflowSingle

	case IntentGreeting:
		// 打招呼：使用第一个 Agent 简单回复
		result.WorkflowReasoning = "关键词快速匹配：用户打招呼"
		result.SelectionReasoning = "简单问候，使用默认 Agent"
		if len(analysisCtx.AvailableAgents) > 0 {
			result.SelectedAgentIDs = []string{analysisCtx.AvailableAgents[0].ID}
		}

	case IntentOrder:
		// 订单查询：优先选择订单相关 Agent，否则使用默认
		result.WorkflowReasoning = "关键词快速匹配：订单相关查询"
		result.SelectionReasoning = "检测到订单关键词"
		// 尝试找订单相关 Agent
		for _, agent := range analysisCtx.AvailableAgents {
			nameLower := strings.ToLower(agent.Name)
			descLower := strings.ToLower(agent.Description)
			if strings.Contains(nameLower, "订单") || strings.Contains(descLower, "订单") ||
				strings.Contains(nameLower, "order") || strings.Contains(descLower, "order") {
				result.SelectedAgentIDs = []string{agent.ID}
				break
			}
		}
		// 未找到专门的订单 Agent，使用默认
		if len(result.SelectedAgentIDs) == 0 && len(analysisCtx.AvailableAgents) > 0 {
			result.SelectedAgentIDs = []string{analysisCtx.AvailableAgents[0].ID}
		}

	default:
		// 其他情况使用默认 Agent
		result.WorkflowReasoning = "关键词匹配"
		result.SelectionReasoning = quickResult.Reason
		if len(analysisCtx.AvailableAgents) > 0 {
			result.SelectedAgentIDs = []string{analysisCtx.AvailableAgents[0].ID}
		}
	}

	return result
}

// GetIntentRouter 获取意图路由器（用于添加自定义规则）
func (qa *QueryAnalyzer) GetIntentRouter() *IntentRouter {
	return qa.intentRouter
}
