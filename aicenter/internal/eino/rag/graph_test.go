package rag

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/cloudwego/eino-ext/components/model/openai"
)

func TestHTTPRetriever(t *testing.T) {
	ragURL := os.Getenv("RAG_URL")
	if ragURL == "" {
		ragURL = "http://localhost:8082"
	}

	collectionID := os.Getenv("TEST_COLLECTION_ID")
	if collectionID == "" {
		collectionID = "c447c20d-a591-4253-84e2-504aeeb7492a"
	}

	retriever, err := NewHTTPRetriever(&HTTPRetrieverConfig{
		BaseURL:      ragURL,
		CollectionID: collectionID,
		TopK:         3,
	})
	if err != nil {
		t.Fatalf("NewHTTPRetriever failed: %v", err)
	}

	ctx := context.Background()
	docs, err := retriever.Retrieve(ctx, "A/B测试")
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	t.Logf("Retrieved %d documents", len(docs))
	for i, doc := range docs {
		t.Logf("[%d] ID=%s, Score=%.4f, Content=%s",
			i+1, doc.ID, doc.Score(), truncate(doc.Content, 100))
	}

	if len(docs) == 0 {
		t.Error("Expected at least one document")
	}
}

func TestRAGGraph(t *testing.T) {
	// Skip if no API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	modelName := os.Getenv("OPENAI_MODEL_NAME")
	if modelName == "" {
		modelName = "gpt-3.5-turbo"
	}

	ragURL := os.Getenv("RAG_URL")
	if ragURL == "" {
		ragURL = "http://localhost:8082"
	}

	collectionID := os.Getenv("TEST_COLLECTION_ID")
	if collectionID == "" {
		collectionID = "c447c20d-a591-4253-84e2-504aeeb7492a"
	}

	ctx := context.Background()

	// 1. 创建检索器
	retriever, err := NewHTTPRetriever(&HTTPRetrieverConfig{
		BaseURL:      ragURL,
		CollectionID: collectionID,
		TopK:         5,
	})
	if err != nil {
		t.Fatalf("NewHTTPRetriever failed: %v", err)
	}

	// 2. 创建 ChatModel
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   modelName,
	})
	if err != nil {
		t.Fatalf("NewChatModel failed: %v", err)
	}

	// 3. 创建 RAG Graph
	ragGraph, err := NewRAGGraph(ctx, &RAGGraphConfig{
		Retriever: retriever,
		ChatModel: chatModel,
		TopK:      5,
	})
	if err != nil {
		t.Fatalf("NewRAGGraph failed: %v", err)
	}

	// 4. 执行查询
	result, err := ragGraph.Run(ctx, "如何做A/B测试？")
	if err != nil {
		t.Fatalf("RAGGraph.Run failed: %v", err)
	}

	t.Logf("RAG Result: %s", result.Content)

	if result.Content == "" {
		t.Error("Expected non-empty content")
	}
}

func TestMultiQueryRetriever(t *testing.T) {
	// Skip if no API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	modelName := os.Getenv("OPENAI_MODEL_NAME")
	if modelName == "" {
		modelName = "gpt-3.5-turbo"
	}

	ragURL := os.Getenv("RAG_URL")
	if ragURL == "" {
		ragURL = "http://localhost:8082"
	}

	collectionID := os.Getenv("TEST_COLLECTION_ID")
	if collectionID == "" {
		collectionID = "c447c20d-a591-4253-84e2-504aeeb7492a"
	}

	ctx := context.Background()

	// 1. 创建基础检索器
	baseRetriever, err := NewHTTPRetriever(&HTTPRetrieverConfig{
		BaseURL:      ragURL,
		CollectionID: collectionID,
		TopK:         3,
	})
	if err != nil {
		t.Fatalf("NewHTTPRetriever failed: %v", err)
	}

	// 2. 创建 LLM
	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   modelName,
	})
	if err != nil {
		t.Fatalf("NewChatModel failed: %v", err)
	}

	// 3. 创建 MultiQuery 检索器
	mqRetriever, err := NewMultiQueryRetriever(ctx, &MultiQueryRetrieverConfig{
		BaseRetriever: baseRetriever,
		RewriteLLM:    llm,
		MaxQueries:    3,
		FusionFunc:    RRFFusionFunc(60),
	})
	if err != nil {
		t.Fatalf("NewMultiQueryRetriever failed: %v", err)
	}

	// 4. 执行检索
	docs, err := mqRetriever.Retrieve(ctx, "测试自动化")
	if err != nil {
		t.Fatalf("MultiQueryRetriever.Retrieve failed: %v", err)
	}

	t.Logf("MultiQuery Retrieved %d documents", len(docs))
	for i, doc := range docs {
		t.Logf("[%d] ID=%s, Content=%s", i+1, doc.ID, truncate(doc.Content, 100))
	}
}

// 简单集成测试
func TestRAGIntegration(t *testing.T) {
	ragURL := os.Getenv("RAG_URL")
	if ragURL == "" {
		ragURL = "http://localhost:8082"
	}

	// 测试 RAG 服务是否可用
	retriever, err := NewHTTPRetriever(&HTTPRetrieverConfig{
		BaseURL:      ragURL,
		CollectionID: "c447c20d-a591-4253-84e2-504aeeb7492a",
		TopK:         3,
	})
	if err != nil {
		t.Fatalf("NewHTTPRetriever failed: %v", err)
	}

	ctx := context.Background()

	// 测试不同查询
	queries := []string{
		"A/B测试",
		"自动化测试",
		"API文档",
	}

	for _, query := range queries {
		t.Run(query, func(t *testing.T) {
			docs, err := retriever.Retrieve(ctx, query)
			if err != nil {
				t.Errorf("Retrieve failed: %v", err)
				return
			}
			t.Logf("Query '%s' returned %d docs", query, len(docs))
			for _, doc := range docs {
				fmt.Printf("  - %.4f: %s\n", doc.Score(), truncate(doc.Content, 80))
			}
		})
	}
}
