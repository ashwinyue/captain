package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/tgo/captain/aicenter/pkg/external/rag"
)

// RAGRetrieveTool provides document retrieval capability
type RAGRetrieveTool struct {
	client       *rag.Client
	collectionID string
	toolInfo     *schema.ToolInfo
}

// NewRAGRetrieveTool creates a new RAG retrieval tool for a specific collection
func NewRAGRetrieveTool(client *rag.Client, collectionID, collectionName string) *RAGRetrieveTool {
	return &RAGRetrieveTool{
		client:       client,
		collectionID: collectionID,
		toolInfo: &schema.ToolInfo{
			Name: fmt.Sprintf("search_%s", sanitizeName(collectionName)),
			Desc: fmt.Sprintf("Search documents in the '%s' knowledge base. Use this to find relevant information.", collectionName),
			ParamsOneOf: schema.NewParamsOneOfByParams(
				map[string]*schema.ParameterInfo{
					"query": {
						Type:     schema.String,
						Desc:     "The search query to find relevant documents",
						Required: true,
					},
					"top_k": {
						Type: schema.Integer,
						Desc: "Number of documents to retrieve (default: 5)",
					},
				},
			),
		},
	}
}

func (t *RAGRetrieveTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return t.toolInfo, nil
}

type ragInput struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k"`
}

func (t *RAGRetrieveTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input ragInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("parse arguments: %w", err)
	}

	if input.TopK == 0 {
		input.TopK = 5
	}

	log.Printf("[RAG Tool] Query: %s, CollectionID: %s", input.Query, t.collectionID)

	resp, err := t.client.Retrieve(ctx, &rag.RetrieveRequest{
		CollectionID: t.collectionID,
		Query:        input.Query,
		TopK:         input.TopK,
	})
	if err != nil {
		log.Printf("[RAG Tool] Error: %v", err)
		return "", fmt.Errorf("retrieve documents: %w", err)
	}

	log.Printf("[RAG Tool] Found %d documents", len(resp.Documents))

	// Format results
	var sb strings.Builder
	if len(resp.Documents) == 0 {
		sb.WriteString("No relevant documents found in knowledge base.")
	} else {
		sb.WriteString(fmt.Sprintf("Found %d relevant documents:\n\n", len(resp.Documents)))
		for i, doc := range resp.Documents {
			sb.WriteString(fmt.Sprintf("--- Document %d (score: %.3f) ---\n", i+1, doc.Score))
			sb.WriteString(doc.Content)
			sb.WriteString("\n\n")
		}
	}

	result := sb.String()
	log.Printf("[RAG Tool] Result: %s", result[:min(200, len(result))])
	return result, nil
}

// LoadRAGTools creates retrieval tools for specified collections
func LoadRAGTools(ctx context.Context, ragURL string, collectionIDs []string) ([]tool.BaseTool, error) {
	if ragURL == "" || len(collectionIDs) == 0 {
		return nil, nil
	}

	client := rag.NewClient(ragURL)

	// Create tools directly using collection IDs as names
	// This avoids needing to call ListCollections which requires project_id
	tools := make([]tool.BaseTool, 0, len(collectionIDs))
	for i, id := range collectionIDs {
		// Use a generic name like "knowledge_base_1", "knowledge_base_2" etc.
		name := fmt.Sprintf("knowledge_base_%d", i+1)
		tools = append(tools, NewRAGRetrieveTool(client, id, name))
	}

	return tools, nil
}

func sanitizeName(name string) string {
	// Replace spaces and special chars with underscores
	result := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, name)
	return strings.ToLower(result)
}
