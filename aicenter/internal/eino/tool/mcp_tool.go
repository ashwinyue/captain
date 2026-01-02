package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/tgo/captain/aicenter/pkg/external/mcp"
)

// MCPToolAdapter wraps an MCP tool as an Eino BaseTool
type MCPToolAdapter struct {
	client   *mcp.Client
	mcpTool  mcp.Tool
	toolInfo *schema.ToolInfo
}

// NewMCPToolAdapter creates a new MCP tool adapter
func NewMCPToolAdapter(client *mcp.Client, mcpTool mcp.Tool) *MCPToolAdapter {
	return &MCPToolAdapter{
		client:  client,
		mcpTool: mcpTool,
		toolInfo: &schema.ToolInfo{
			Name: mcpTool.Name,
			Desc: mcpTool.Description,
			ParamsOneOf: schema.NewParamsOneOfByParams(
				map[string]*schema.ParameterInfo{
					"input": {
						Type: schema.String,
						Desc: "Tool input parameters as JSON",
					},
				},
			),
		},
	}
}

func (t *MCPToolAdapter) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return t.toolInfo, nil
}

func (t *MCPToolAdapter) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input map[string]interface{}
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("parse arguments: %w", err)
	}

	resp, err := t.client.CallTool(ctx, &mcp.ToolCallRequest{
		Name:  t.mcpTool.Name,
		Input: input,
	})
	if err != nil {
		return "", err
	}

	if resp.IsError {
		return "", fmt.Errorf("tool error: %s", resp.Content)
	}

	return resp.Content, nil
}

// LoadMCPTools loads all tools from an MCP server
func LoadMCPTools(ctx context.Context, mcpURL string) ([]tool.BaseTool, error) {
	if mcpURL == "" {
		return nil, nil
	}

	// Detect transport type based on URL
	var tools []tool.BaseTool

	if strings.Contains(mcpURL, "/sse") || strings.HasSuffix(mcpURL, ":sse") {
		// Use SSE client
		sseClient := mcp.NewSSEClient(mcpURL)
		mcpTools, err := sseClient.ListTools(ctx)
		if err != nil {
			return nil, fmt.Errorf("list MCP tools (SSE): %w", err)
		}

		for _, t := range mcpTools {
			tools = append(tools, NewMCPSSEToolAdapter(sseClient, t))
		}
	} else {
		// Use HTTP client
		httpClient := mcp.NewClient(mcpURL)
		mcpTools, err := httpClient.ListTools(ctx)
		if err != nil {
			return nil, fmt.Errorf("list MCP tools (HTTP): %w", err)
		}

		for _, t := range mcpTools {
			tools = append(tools, NewMCPToolAdapter(httpClient, t))
		}
	}

	return tools, nil
}

// MCPSSEToolAdapter wraps an MCP SSE tool as an Eino BaseTool
type MCPSSEToolAdapter struct {
	client   *mcp.SSEClient
	mcpTool  mcp.MCPTool
	toolInfo *schema.ToolInfo
}

// NewMCPSSEToolAdapter creates a new MCP SSE tool adapter
func NewMCPSSEToolAdapter(client *mcp.SSEClient, mcpTool mcp.MCPTool) *MCPSSEToolAdapter {
	return &MCPSSEToolAdapter{
		client:  client,
		mcpTool: mcpTool,
		toolInfo: &schema.ToolInfo{
			Name: mcpTool.Name,
			Desc: mcpTool.Description,
			ParamsOneOf: schema.NewParamsOneOfByParams(
				map[string]*schema.ParameterInfo{
					"input": {
						Type: schema.String,
						Desc: "Tool input parameters as JSON",
					},
				},
			),
		},
	}
}

func (t *MCPSSEToolAdapter) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return t.toolInfo, nil
}

func (t *MCPSSEToolAdapter) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input map[string]interface{}
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("parse arguments: %w", err)
	}

	result, err := t.client.CallTool(ctx, t.mcpTool.Name, input)
	if err != nil {
		return "", err
	}

	if result.IsError {
		if len(result.Content) > 0 {
			return "", fmt.Errorf("tool error: %s", result.Content[0].Text)
		}
		return "", fmt.Errorf("tool error")
	}

	// Combine all text content
	var output strings.Builder
	for _, block := range result.Content {
		if block.Type == "text" {
			output.WriteString(block.Text)
		}
	}

	return output.String(), nil
}
