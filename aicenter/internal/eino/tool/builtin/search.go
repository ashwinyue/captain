package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// DuckDuckGoSearchTool provides web search capability
type DuckDuckGoSearchTool struct {
	httpClient *http.Client
	toolInfo   *schema.ToolInfo
}

func NewDuckDuckGoSearchTool() *DuckDuckGoSearchTool {
	return &DuckDuckGoSearchTool{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		toolInfo: &schema.ToolInfo{
			Name: "web_search",
			Desc: "Search the web for information using DuckDuckGo. Use this to find current information about topics.",
			ParamsOneOf: schema.NewParamsOneOfByParams(
				map[string]*schema.ParameterInfo{
					"query": {
						Type:     schema.String,
						Desc:     "The search query",
						Required: true,
					},
				},
			),
		},
	}
}

func (t *DuckDuckGoSearchTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return t.toolInfo, nil
}

type searchInput struct {
	Query string `json:"query"`
}

func (t *DuckDuckGoSearchTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input searchInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("parse arguments: %w", err)
	}

	// Use DuckDuckGo instant answer API
	apiURL := "https://api.duckduckgo.com/?q=" + url.QueryEscape(input.Query) + "&format=json&no_html=1"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	// Extract relevant information
	var output string

	if abstract, ok := result["Abstract"].(string); ok && abstract != "" {
		output += fmt.Sprintf("Summary: %s\n\n", abstract)
	}

	if answer, ok := result["Answer"].(string); ok && answer != "" {
		output += fmt.Sprintf("Answer: %s\n\n", answer)
	}

	if relatedTopics, ok := result["RelatedTopics"].([]interface{}); ok {
		output += "Related Topics:\n"
		for i, topic := range relatedTopics {
			if i >= 5 {
				break
			}
			if t, ok := topic.(map[string]interface{}); ok {
				if text, ok := t["Text"].(string); ok {
					output += fmt.Sprintf("- %s\n", text)
				}
			}
		}
	}

	if output == "" {
		output = "No results found for the query."
	}

	return output, nil
}
