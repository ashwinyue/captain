package rag

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

// MultiQueryRetrieverConfig MultiQuery 检索器配置
type MultiQueryRetrieverConfig struct {
	BaseRetriever retriever.Retriever // 基础检索器
	RewriteLLM    model.ChatModel     // 用于重写查询的 LLM
	MaxQueries    int                 // 最大查询数量
	FusionFunc    FusionFunc          // 结果融合函数
}

// FusionFunc 结果融合函数类型
type FusionFunc func(results [][]*schema.Document) []*schema.Document

// MultiQueryRetriever 多查询检索器
// 使用 LLM 将原始查询重写为多个角度的查询，然后合并结果
type MultiQueryRetriever struct {
	config        *MultiQueryRetrieverConfig
	rewritePrompt prompt.ChatTemplate
}

// NewMultiQueryRetriever 创建多查询检索器
func NewMultiQueryRetriever(ctx context.Context, config *MultiQueryRetrieverConfig) (*MultiQueryRetriever, error) {
	if config.BaseRetriever == nil {
		return nil, fmt.Errorf("base retriever is required")
	}
	if config.RewriteLLM == nil {
		return nil, fmt.Errorf("rewrite LLM is required")
	}
	if config.MaxQueries <= 0 {
		config.MaxQueries = 3
	}
	if config.FusionFunc == nil {
		config.FusionFunc = defaultFusionFunc
	}

	// 创建重写提示模板
	rewritePrompt := prompt.FromMessages(schema.FString,
		schema.SystemMessage(multiQuerySystemPrompt),
		schema.UserMessage(multiQueryUserPrompt),
	)

	return &MultiQueryRetriever{
		config:        config,
		rewritePrompt: rewritePrompt,
	}, nil
}

// Retrieve 实现 retriever.Retriever 接口
func (r *MultiQueryRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	log.Printf("[MultiQueryRetriever] Original query: %s", truncate(query, 50))

	// 1. 使用 LLM 重写查询
	queries, err := r.rewriteQuery(ctx, query)
	if err != nil {
		log.Printf("[MultiQueryRetriever] Rewrite failed, using original: %v", err)
		queries = []string{query}
	}

	log.Printf("[MultiQueryRetriever] Generated %d queries", len(queries))

	// 2. 对每个查询执行检索
	allResults := make([][]*schema.Document, 0, len(queries))
	for i, q := range queries {
		docs, err := r.config.BaseRetriever.Retrieve(ctx, q, opts...)
		if err != nil {
			log.Printf("[MultiQueryRetriever] Query %d failed: %v", i, err)
			continue
		}
		allResults = append(allResults, docs)
	}

	// 3. 融合结果
	fusedDocs := r.config.FusionFunc(allResults)
	log.Printf("[MultiQueryRetriever] Fused to %d unique documents", len(fusedDocs))

	return fusedDocs, nil
}

// rewriteQuery 使用 LLM 重写查询
func (r *MultiQueryRetriever) rewriteQuery(ctx context.Context, query string) ([]string, error) {
	// 构建提示
	messages, err := r.rewritePrompt.Format(ctx, map[string]any{
		"query":       query,
		"num_queries": r.config.MaxQueries,
	})
	if err != nil {
		return nil, fmt.Errorf("format prompt: %w", err)
	}

	// 调用 LLM
	resp, err := r.config.RewriteLLM.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("generate: %w", err)
	}

	// 解析结果（每行一个查询）
	lines := strings.Split(strings.TrimSpace(resp.Content), "\n")
	queries := make([]string, 0, len(lines)+1)
	queries = append(queries, query) // 保留原始查询

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 移除序号前缀 (1. 2. 3. 等)
		if len(line) > 2 && line[1] == '.' {
			line = strings.TrimSpace(line[2:])
		}
		if len(line) > 3 && line[2] == '.' {
			line = strings.TrimSpace(line[3:])
		}
		if line != "" && line != query {
			queries = append(queries, line)
		}
	}

	// 限制数量
	if len(queries) > r.config.MaxQueries {
		queries = queries[:r.config.MaxQueries]
	}

	return queries, nil
}

// GetType 返回检索器类型
func (r *MultiQueryRetriever) GetType() string {
	return "MultiQueryRetriever"
}

// IsCallbacksEnabled 是否启用回调
func (r *MultiQueryRetriever) IsCallbacksEnabled() bool {
	return true
}

// defaultFusionFunc 默认融合函数：去重并按分数排序
func defaultFusionFunc(results [][]*schema.Document) []*schema.Document {
	seen := make(map[string]bool)
	var merged []*schema.Document

	for _, docs := range results {
		for _, doc := range docs {
			if !seen[doc.ID] {
				seen[doc.ID] = true
				merged = append(merged, doc)
			}
		}
	}

	return merged
}

// RRFFusionFunc Reciprocal Rank Fusion 融合函数
func RRFFusionFunc(k int) FusionFunc {
	if k <= 0 {
		k = 60 // 默认 RRF 常数
	}

	return func(results [][]*schema.Document) []*schema.Document {
		scores := make(map[string]float64)
		docs := make(map[string]*schema.Document)

		for _, docList := range results {
			for rank, doc := range docList {
				// RRF 分数: 1 / (k + rank)
				score := 1.0 / float64(k+rank+1)
				scores[doc.ID] += score
				if _, exists := docs[doc.ID]; !exists {
					docs[doc.ID] = doc
				}
			}
		}

		// 按 RRF 分数排序
		type scoredDoc struct {
			doc   *schema.Document
			score float64
		}
		var sorted []scoredDoc
		for id, doc := range docs {
			sorted = append(sorted, scoredDoc{doc: doc, score: scores[id]})
		}

		// 简单冒泡排序（文档数量通常不多）
		for i := 0; i < len(sorted)-1; i++ {
			for j := 0; j < len(sorted)-i-1; j++ {
				if sorted[j].score < sorted[j+1].score {
					sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
				}
			}
		}

		result := make([]*schema.Document, len(sorted))
		for i, sd := range sorted {
			result[i] = sd.doc
		}

		return result
	}
}

const multiQuerySystemPrompt = `你是一个查询重写专家。你的任务是将用户的查询从多个角度重写，以获得更全面的搜索结果。

规则：
1. 生成 {num_queries} 个不同角度的查询
2. 每个查询一行
3. 保持查询简洁，不要添加无关内容
4. 不要重复原始查询`

const multiQueryUserPrompt = `原始查询: {query}

请从不同角度重写这个查询，每行一个：`
