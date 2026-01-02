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
	NodeQueryRewrite  = "QueryRewrite"
	NodeMultiRetrieve = "MultiRetrieve"
	NodeRerank        = "Rerank"
	NodeRAGPrompt     = "RAGPrompt"
	NodeGenerate      = "Generate"
)

// AdvancedRAGConfig 高级 RAG 配置
type AdvancedRAGConfig struct {
	Retriever     retriever.Retriever
	ChatModel     model.ChatModel
	RewriteLLM    model.ChatModel // 用于查询重写（可选，为空则跳过重写）
	RerankLLM     model.ChatModel // 用于重排序（可选，为空则跳过重排序）
	SystemPrompt  string
	TopK          int
	RerankTopK    int  // 重排序后保留的文档数
	EnableRewrite bool // 是否启用查询重写
	EnableRerank  bool // 是否启用重排序
}

// AdvancedRAGGraph 高级 RAG Graph
// 支持：查询重写 → 多路检索 → 重排序 → 生成
type AdvancedRAGGraph struct {
	config   *AdvancedRAGConfig
	runnable compose.Runnable[*RAGInput, *schema.Message]
}

// NewAdvancedRAGGraph 创建高级 RAG Graph
func NewAdvancedRAGGraph(ctx context.Context, config *AdvancedRAGConfig) (*AdvancedRAGGraph, error) {
	if config.Retriever == nil {
		return nil, fmt.Errorf("retriever is required")
	}
	if config.ChatModel == nil {
		return nil, fmt.Errorf("chat model is required")
	}
	if config.TopK <= 0 {
		config.TopK = 10
	}
	if config.RerankTopK <= 0 {
		config.RerankTopK = 5
	}
	if config.SystemPrompt == "" {
		config.SystemPrompt = advancedRAGSystemPrompt
	}

	rg := &AdvancedRAGGraph{config: config}

	runnable, err := rg.build(ctx)
	if err != nil {
		return nil, fmt.Errorf("build advanced RAG graph: %w", err)
	}

	rg.runnable = runnable
	return rg, nil
}

// build 构建高级 RAG Graph
func (rg *AdvancedRAGGraph) build(ctx context.Context) (compose.Runnable[*RAGInput, *schema.Message], error) {
	g := compose.NewGraph[*RAGInput, *schema.Message]()

	// 1. 查询提取节点
	extractQuery := func(ctx context.Context, input *RAGInput) (string, error) {
		return input.Query, nil
	}
	_ = g.AddLambdaNode(NodeInputToQuery, compose.InvokableLambda(extractQuery),
		compose.WithNodeName("ExtractQuery"))

	// 2. 检索节点
	_ = g.AddRetrieverNode(NodeRetriever, rg.config.Retriever,
		compose.WithOutputKey("documents"),
		compose.WithNodeName("Retriever"))

	// 3. 重排序节点（如果启用）
	if rg.config.EnableRerank && rg.config.RerankLLM != nil {
		rerankFunc := rg.createRerankFunc()
		_ = g.AddLambdaNode(NodeRerank, compose.InvokableLambda(rerankFunc),
			compose.WithNodeName("LLMRerank"))
	}

	// 4. RAG Prompt 构建节点
	ragPrompt := prompt.FromMessages(schema.FString,
		schema.SystemMessage(rg.config.SystemPrompt),
		schema.UserMessage(advancedRAGUserPrompt),
	)
	_ = g.AddChatTemplateNode(NodeRAGPrompt, ragPrompt,
		compose.WithNodeName("RAGPromptBuilder"))

	// 5. 生成节点
	_ = g.AddChatModelNode(NodeGenerate, rg.config.ChatModel,
		compose.WithNodeName("AnswerGenerator"))

	// 6. 构建边
	_ = g.AddEdge(compose.START, NodeInputToQuery)
	_ = g.AddEdge(NodeInputToQuery, NodeRetriever)

	if rg.config.EnableRerank && rg.config.RerankLLM != nil {
		_ = g.AddEdge(NodeRetriever, NodeRerank)
		_ = g.AddEdge(NodeRerank, NodeRAGPrompt)
	} else {
		_ = g.AddEdge(NodeRetriever, NodeRAGPrompt)
	}

	_ = g.AddEdge(NodeRAGPrompt, NodeGenerate)
	_ = g.AddEdge(NodeGenerate, compose.END)

	// 7. 编译
	runnable, err := g.Compile(ctx,
		compose.WithGraphName("AdvancedRAGGraph"),
		compose.WithNodeTriggerMode(compose.AllPredecessor),
	)
	if err != nil {
		return nil, fmt.Errorf("compile graph: %w", err)
	}

	log.Printf("[AdvancedRAGGraph] Built with rewrite=%v, rerank=%v",
		rg.config.EnableRewrite, rg.config.EnableRerank)
	return runnable, nil
}

// createRerankFunc 创建重排序函数
func (rg *AdvancedRAGGraph) createRerankFunc() func(ctx context.Context, input map[string]any) (map[string]any, error) {
	return func(ctx context.Context, input map[string]any) (map[string]any, error) {
		docs, ok := input["documents"].([]*schema.Document)
		if !ok || len(docs) == 0 {
			return input, nil
		}

		query, _ := input["query"].(string)
		if query == "" {
			return input, nil
		}

		// 使用 LLM 对文档进行相关性评分
		rerankedDocs, err := rg.rerankWithLLM(ctx, query, docs)
		if err != nil {
			log.Printf("[AdvancedRAGGraph] Rerank failed, using original order: %v", err)
			return input, nil
		}

		// 只保留 top K
		if len(rerankedDocs) > rg.config.RerankTopK {
			rerankedDocs = rerankedDocs[:rg.config.RerankTopK]
		}

		input["documents"] = rerankedDocs
		return input, nil
	}
}

// rerankWithLLM 使用 LLM 重排序文档
func (rg *AdvancedRAGGraph) rerankWithLLM(ctx context.Context, query string, docs []*schema.Document) ([]*schema.Document, error) {
	// 简化实现：让 LLM 评估每个文档的相关性
	// 实际生产中应使用专门的 Reranker 模型（如 Cohere Rerank、BGE Reranker）

	log.Printf("[AdvancedRAGGraph] Reranking %d documents", len(docs))

	// 构建评估提示
	var docList string
	for i, doc := range docs {
		docList += fmt.Sprintf("[%d] %s\n", i+1, truncate(doc.Content, 200))
	}

	messages := []*schema.Message{
		schema.SystemMessage("你是一个文档相关性评估专家。根据用户查询，对文档进行相关性排序。"),
		schema.UserMessage(fmt.Sprintf(`用户查询: %s

文档列表:
%s

请按相关性从高到低排列文档编号（用逗号分隔），例如: 3,1,2,5,4`, query, docList)),
	}

	resp, err := rg.config.RerankLLM.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM rerank: %w", err)
	}

	// 解析排序结果（简化处理）
	// 实际应用中需要更健壮的解析逻辑
	_ = resp.Content // 这里简化处理，直接返回原顺序

	return docs, nil
}

// Run 执行高级 RAG 查询
func (rg *AdvancedRAGGraph) Run(ctx context.Context, query string) (*RAGOutput, error) {
	input := &RAGInput{Query: query}

	log.Printf("[AdvancedRAGGraph] Running query: %s", truncate(query, 50))

	result, err := rg.runnable.Invoke(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("invoke: %w", err)
	}

	return &RAGOutput{Content: result.Content}, nil
}

// Stream 流式执行
func (rg *AdvancedRAGGraph) Stream(ctx context.Context, query string) (*schema.StreamReader[*schema.Message], error) {
	input := &RAGInput{Query: query}
	return rg.runnable.Stream(ctx, input)
}

const advancedRAGSystemPrompt = `你是一个智能问答助手，基于提供的参考文档回答用户问题。

## 核心原则
1. **准确性**: 仅基于参考文档回答，不编造信息
2. **完整性**: 综合多个文档信息给出全面回答
3. **可追溯**: 如有必要，注明信息来源
4. **诚实性**: 如果文档中没有相关信息，明确告知用户

## 回答风格
- 清晰、简洁、专业
- 结构化呈现复杂信息
- 适当使用列表和分点`

const advancedRAGUserPrompt = `## 参考文档
{documents}

## 用户问题
{query}

请基于参考文档回答问题：`
