package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// ConflictInfo 冲突信息
type ConflictInfo struct {
	Detected    bool     `json:"detected"`    // 是否检测到冲突
	Description string   `json:"description"` // 冲突描述
	AgentIDs    []string `json:"agent_ids"`   // 冲突的 Agent IDs
	Resolution  string   `json:"resolution"`  // 解决方案
}

// ConsolidationResult 聚合结果（包含冲突检测）
type ConsolidationResult struct {
	Content      string       `json:"content"`       // 最终内容
	Conflict     ConflictInfo `json:"conflict"`      // 冲突信息
	Consensus    bool         `json:"consensus"`     // 是否达成共识
	SourceCount  int          `json:"source_count"`  // 来源数量
	SuccessCount int          `json:"success_count"` // 成功执行数量
}

// ResultConsolidator 结果聚合器，负责将多个 Agent 的结果合并为最终答案
type ResultConsolidator struct {
	chatModel model.ChatModel
}

// NewResultConsolidator 创建结果聚合器
func NewResultConsolidator(chatModel model.ChatModel) *ResultConsolidator {
	return &ResultConsolidator{
		chatModel: chatModel,
	}
}

// 聚合系统提示模板
const consolidationSystemPrompt = `你是一个智能助手，负责将多个子任务的结果整合为一个完整、连贯的回答。

## 要求
1. 整合所有结果，生成一个完整、连贯的回答
2. 去除重复信息
3. 如果某个子任务失败，说明该部分信息暂时无法获取
4. 回答应该直接针对用户的原始问题
5. 不要提及"子任务"、"结果1"等内部概念

请直接输出整合后的回答。`

// 冲突检测系统提示模板
const conflictDetectionSystemPrompt = `你是一个专业的信息分析助手，负责检测多个来源的回答之间是否存在冲突或矛盾。

## 任务
分析以下多个回答，判断它们之间是否存在：
1. 事实性冲突（如数字、日期、名称等不一致）
2. 观点性矛盾（如对同一问题给出相反的建议）
3. 信息不一致（如同一事物的不同描述）

## 输出格式（JSON）
{
  "has_conflict": true/false,
  "conflict_description": "冲突的具体描述（如无冲突则为空）",
  "conflicting_sources": [1, 2],  // 冲突的来源编号
  "resolution_suggestion": "建议的解决方案（如无冲突则为空）",
  "consensus_points": ["共识点1", "共识点2"]  // 各来源的共识点
}

只输出 JSON，不要输出其他内容。`

// 冲突检测用户消息模板
const conflictDetectionUserPrompt = `## 用户原始问题
{original_query}

## 各来源的回答
{results_content}

请分析以上回答是否存在冲突。`

// 聚合用户消息模板 (FString 格式)
const consolidationUserPrompt = `## 用户原始问题
{original_query}

## 各子任务的执行结果
{results_content}`

// Consolidate 聚合多个 Agent 的执行结果
func (rc *ResultConsolidator) Consolidate(ctx context.Context, execResult *ExecutionResult, originalQuery string) (string, error) {
	// 单 Agent 结果直接返回
	if execResult.Workflow == WorkflowSingle || len(execResult.Results) == 1 {
		if len(execResult.Results) > 0 {
			return execResult.Results[0].Content, execResult.Results[0].Error
		}
		return "", fmt.Errorf("no results to consolidate")
	}

	// 多 Agent 结果需要聚合
	log.Printf("[ResultConsolidator] Consolidating %d results", len(execResult.Results))

	// 使用 eino prompt.FromMessages 构建消息模板
	chatTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage(consolidationSystemPrompt),
		schema.UserMessage(consolidationUserPrompt),
	)

	// 构建结果内容
	resultsContent := rc.formatResults(execResult)

	// 格式化模板
	messages, err := chatTpl.Format(ctx, map[string]any{
		"original_query":  originalQuery,
		"results_content": resultsContent,
	})
	if err != nil {
		log.Printf("[ResultConsolidator] Template format failed: %v, using simple merge", err)
		return rc.simpleMerge(execResult), nil
	}

	// 调用 LLM 聚合
	resp, err := rc.chatModel.Generate(ctx, messages)
	if err != nil {
		log.Printf("[ResultConsolidator] LLM consolidation failed: %v, using simple merge", err)
		return rc.simpleMerge(execResult), nil
	}

	log.Printf("[ResultConsolidator] Consolidated result length: %d", len(resp.Content))
	return resp.Content, nil
}

// formatResults 格式化执行结果
func (rc *ResultConsolidator) formatResults(execResult *ExecutionResult) string {
	var sb strings.Builder

	for i, result := range execResult.Results {
		if result.Error != nil {
			sb.WriteString(fmt.Sprintf("### 结果 %d (执行失败)\n错误: %v\n\n", i+1, result.Error))
		} else {
			subQuestion := ""
			if result.SubQuestion != nil {
				subQuestion = fmt.Sprintf("\n子问题: %s", result.SubQuestion.Question)
			}
			sb.WriteString(fmt.Sprintf("### 结果 %d%s\n%s\n\n", i+1, subQuestion, result.Content))
		}
	}

	return sb.String()
}

// simpleMerge 简单合并结果（降级策略）
func (rc *ResultConsolidator) simpleMerge(execResult *ExecutionResult) string {
	var parts []string

	for _, result := range execResult.Results {
		if result.Error == nil && result.Content != "" {
			parts = append(parts, result.Content)
		}
	}

	if len(parts) == 0 {
		return "抱歉，无法获取相关信息。"
	}

	return strings.Join(parts, "\n\n---\n\n")
}

// ConsolidateSimple 简单聚合（不使用 LLM）
func (rc *ResultConsolidator) ConsolidateSimple(execResult *ExecutionResult) string {
	return rc.simpleMerge(execResult)
}

// ConsolidateWithConflictDetection 带冲突检测的聚合
func (rc *ResultConsolidator) ConsolidateWithConflictDetection(ctx context.Context, execResult *ExecutionResult, originalQuery string) (*ConsolidationResult, error) {
	successCount := 0
	for _, r := range execResult.Results {
		if r.Error == nil {
			successCount++
		}
	}

	// 单结果无需冲突检测
	if len(execResult.Results) <= 1 || successCount <= 1 {
		content := ""
		if len(execResult.Results) > 0 && execResult.Results[0].Error == nil {
			content = execResult.Results[0].Content
		}
		return &ConsolidationResult{
			Content:      content,
			Conflict:     ConflictInfo{Detected: false},
			Consensus:    true,
			SourceCount:  len(execResult.Results),
			SuccessCount: successCount,
		}, nil
	}

	// 检测冲突
	conflict, err := rc.detectConflicts(ctx, execResult, originalQuery)
	if err != nil {
		log.Printf("[ResultConsolidator] Conflict detection failed: %v, skipping", err)
		conflict = ConflictInfo{Detected: false}
	}

	// 执行聚合
	content, err := rc.Consolidate(ctx, execResult, originalQuery)
	if err != nil {
		return nil, err
	}

	return &ConsolidationResult{
		Content:      content,
		Conflict:     conflict,
		Consensus:    !conflict.Detected,
		SourceCount:  len(execResult.Results),
		SuccessCount: successCount,
	}, nil
}

// detectConflicts 检测多个结果之间的冲突
func (rc *ResultConsolidator) detectConflicts(ctx context.Context, execResult *ExecutionResult, originalQuery string) (ConflictInfo, error) {
	// 使用 eino prompt 构建检测消息
	chatTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage(conflictDetectionSystemPrompt),
		schema.UserMessage(conflictDetectionUserPrompt),
	)

	resultsContent := rc.formatResults(execResult)

	messages, err := chatTpl.Format(ctx, map[string]any{
		"original_query":  originalQuery,
		"results_content": resultsContent,
	})
	if err != nil {
		return ConflictInfo{}, err
	}

	// 调用 LLM 检测冲突
	resp, err := rc.chatModel.Generate(ctx, messages)
	if err != nil {
		return ConflictInfo{}, err
	}

	// 解析 JSON 响应
	return rc.parseConflictResponse(resp.Content, execResult)
}

// parseConflictResponse 解析冲突检测响应
func (rc *ResultConsolidator) parseConflictResponse(content string, execResult *ExecutionResult) (ConflictInfo, error) {
	// 简单解析 JSON 响应
	content = strings.TrimSpace(content)

	// 移除可能的 markdown 代码块
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	// 解析 JSON
	var result struct {
		HasConflict          bool     `json:"has_conflict"`
		ConflictDescription  string   `json:"conflict_description"`
		ConflictingSources   []int    `json:"conflicting_sources"`
		ResolutionSuggestion string   `json:"resolution_suggestion"`
		ConsensusPoints      []string `json:"consensus_points"`
	}

	if err := json.Unmarshal([]byte(content), &result); err != nil {
		log.Printf("[ResultConsolidator] Failed to parse conflict response: %v", err)
		return ConflictInfo{Detected: false}, nil
	}

	// 转换 source 索引为 agent ID
	var agentIDs []string
	for _, idx := range result.ConflictingSources {
		if idx > 0 && idx <= len(execResult.Results) {
			agentIDs = append(agentIDs, execResult.Results[idx-1].AgentID)
		}
	}

	return ConflictInfo{
		Detected:    result.HasConflict,
		Description: result.ConflictDescription,
		AgentIDs:    agentIDs,
		Resolution:  result.ResolutionSuggestion,
	}, nil
}
