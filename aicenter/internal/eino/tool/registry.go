package tool

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
)

// Registry manages tool loading and caching
type Registry struct {
	mcpURL string
	ragURL string
}

// NewRegistry creates a new tool registry
func NewRegistry(mcpURL, ragURL string) *Registry {
	return &Registry{
		mcpURL: mcpURL,
		ragURL: ragURL,
	}
}

// LoadToolsRequest specifies which tools to load
type LoadToolsRequest struct {
	MCPURL        string   // Override default MCP URL
	RAGURL        string   // Override default RAG URL
	CollectionIDs []string // RAG collections to enable
	EnableMCP     bool     // Whether to load MCP tools
}

// LoadTools loads tools based on the request
func (r *Registry) LoadTools(ctx context.Context, req *LoadToolsRequest) ([]tool.BaseTool, error) {
	var allTools []tool.BaseTool

	// Determine URLs
	mcpURL := r.mcpURL
	if req.MCPURL != "" {
		mcpURL = req.MCPURL
	}

	ragURL := r.ragURL
	if req.RAGURL != "" {
		ragURL = req.RAGURL
	}

	// Load MCP tools
	if req.EnableMCP && mcpURL != "" {
		mcpTools, err := LoadMCPTools(ctx, mcpURL)
		if err != nil {
			return nil, fmt.Errorf("load MCP tools: %w", err)
		}
		allTools = append(allTools, mcpTools...)
	}

	// Load RAG tools
	if len(req.CollectionIDs) > 0 && ragURL != "" {
		ragTools, err := LoadRAGTools(ctx, ragURL, req.CollectionIDs)
		if err != nil {
			return nil, fmt.Errorf("load RAG tools: %w", err)
		}
		allTools = append(allTools, ragTools...)
	}

	return allTools, nil
}
