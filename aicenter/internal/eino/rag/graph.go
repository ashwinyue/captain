package rag

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	NodeInputToQuery = "InputToQuery"
	NodeRetriever    = "Retriever"
	NodeChatTemplate = "ChatTemplate"
	NodeChatModel    = "ChatModel"
)

// RAGInput RAG 输入
type RAGInput struct {
	Query   string            `json:"query"`
	History []*schema.Message `json:"history,omitempty"`
}

// RAGOutput RAG 输出
type RAGOutput struct {
	Content   string             `json:"content"`
	Documents []*schema.Document `json:"documents,omitempty"`
}

// RAGGraphConfig RAG Graph 配置
type RAGGraphConfig struct {
	Retriever     retriever.Retriever
	ChatModel     model.ChatModel
	SystemPrompt  string
	TopK          int
	EnableHistory bool
}

// RAGGraph 基于 eino compose.Graph 的 RAG 实现
type RAGGraph struct {
	config   *RAGGraphConfig
	runnable compose.Runnable[*RAGInput, *schema.Message]
}

// NewRAGGraph 创建 RAG Graph
func NewRAGGraph(ctx context.Context, config *RAGGraphConfig) (*RAGGraph, error) {
	if config.Retriever == nil {
		return nil, fmt.Errorf("retriever is required")
	}
	if config.ChatModel == nil {
		return nil, fmt.Errorf("chat model is required")
	}
	if config.TopK <= 0 {
		config.TopK = 5
	}
	if config.SystemPrompt == "" {
		config.SystemPrompt = defaultRAGSystemPrompt
	}

	rg := &RAGGraph{config: config}

	runnable, err := rg.build(ctx)
	if err != nil {
		return nil, fmt.Errorf("build RAG graph: %w", err)
	}

	rg.runnable = runnable
	return rg, nil
}

// build 构建 RAG Graph
func (rg *RAGGraph) build(ctx context.Context) (compose.Runnable[*RAGInput, *schema.Message], error) {
	g := compose.NewGraph[*RAGInput, *schema.Message]()

	// 1. 添加 InputToQuery 节点：提取查询
	inputToQuery := func(ctx context.Context, input *RAGInput) (string, error) {
		return input.Query, nil
	}
	_ = g.AddLambdaNode(NodeInputToQuery, compose.InvokableLambda(inputToQuery),
		compose.WithNodeName("ExtractQuery"))

	// 2. 添加 Retriever 节点：向量检索
	_ = g.AddRetrieverNode(NodeRetriever, rg.config.Retriever,
		compose.WithOutputKey("documents"),
		compose.WithNodeName("VectorRetriever"))

	// 3. 添加 ChatTemplate 节点：构建 RAG Prompt
	chatTpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage(rg.config.SystemPrompt),
		schema.UserMessage(ragUserPromptTemplate),
	)
	_ = g.AddChatTemplateNode(NodeChatTemplate, chatTpl,
		compose.WithNodeName("RAGPromptBuilder"))

	// 4. 添加 ChatModel 节点：生成回答
	_ = g.AddChatModelNode(NodeChatModel, rg.config.ChatModel,
		compose.WithNodeName("AnswerGenerator"))

	// 5. 构建边：定义执行流程
	// START → InputToQuery → Retriever → ChatTemplate → ChatModel → END
	_ = g.AddEdge(compose.START, NodeInputToQuery)
	_ = g.AddEdge(NodeInputToQuery, NodeRetriever)
	_ = g.AddEdge(NodeRetriever, NodeChatTemplate)
	_ = g.AddEdge(NodeChatTemplate, NodeChatModel)
	_ = g.AddEdge(NodeChatModel, compose.END)

	// 6. 编译 Graph
	runnable, err := g.Compile(ctx,
		compose.WithGraphName("RAGGraph"),
		compose.WithNodeTriggerMode(compose.AllPredecessor),
	)
	if err != nil {
		return nil, fmt.Errorf("compile graph: %w", err)
	}

	log.Printf("[RAGGraph] Built successfully with %d nodes", 4)
	return runnable, nil
}

// Run 执行 RAG 查询
func (rg *RAGGraph) Run(ctx context.Context, query string) (*RAGOutput, error) {
	return rg.RunWithHistory(ctx, query, nil)
}

// RunWithHistory 执行 RAG 查询（带历史）
func (rg *RAGGraph) RunWithHistory(ctx context.Context, query string, history []*schema.Message) (*RAGOutput, error) {
	input := &RAGInput{
		Query:   query,
		History: history,
	}

	log.Printf("[RAGGraph] Running query: %s", truncate(query, 50))

	result, err := rg.runnable.Invoke(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("invoke RAG graph: %w", err)
	}

	return &RAGOutput{
		Content: result.Content,
	}, nil
}

// Stream 流式执行 RAG 查询
func (rg *RAGGraph) Stream(ctx context.Context, query string) (*schema.StreamReader[*schema.Message], error) {
	input := &RAGInput{
		Query: query,
	}

	log.Printf("[RAGGraph] Streaming query: %s", truncate(query, 50))

	return rg.runnable.Stream(ctx, input)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// 默认 RAG 系统提示
const defaultRAGSystemPrompt = `你是一个智能助手，根据提供的参考文档回答用户问题。

## 规则
1. 仅基于参考文档中的信息回答问题
2. 如果文档中没有相关信息，明确告知用户
3. 回答准确、简洁、专业
4. 如有必要，引用文档来源`

// RAG 用户提示模板
const ragUserPromptTemplate = `## 参考文档
{documents}

## 用户问题
{query}

请根据参考文档回答用户问题。`
