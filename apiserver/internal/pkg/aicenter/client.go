package aicenter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) request(ctx context.Context, method, path string, body interface{}, headers map[string]string) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}

	return respBody, resp.StatusCode, nil
}

// Agents

func (c *Client) ListAgents(ctx context.Context, projectID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodGet, "/api/v1/agents?project_id="+projectID, nil, headers)
}

func (c *Client) GetAgent(ctx context.Context, agentID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodGet, "/api/v1/agents/"+agentID, nil, headers)
}

func (c *Client) CreateAgent(ctx context.Context, body interface{}, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/api/v1/agents", body, headers)
}

func (c *Client) UpdateAgent(ctx context.Context, agentID string, body interface{}, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodPatch, "/api/v1/agents/"+agentID, body, headers)
}

func (c *Client) DeleteAgent(ctx context.Context, agentID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodDelete, "/api/v1/agents/"+agentID, nil, headers)
}

func (c *Client) RunAgent(ctx context.Context, agentID string, body interface{}, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/api/v1/agents/"+agentID+"/run", body, headers)
}

// Teams

func (c *Client) ListTeams(ctx context.Context, projectID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodGet, "/api/v1/teams?project_id="+projectID, nil, headers)
}

func (c *Client) GetTeam(ctx context.Context, teamID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodGet, "/api/v1/teams/"+teamID, nil, headers)
}

func (c *Client) GetDefaultTeam(ctx context.Context, projectID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodGet, "/api/v1/teams/default?project_id="+projectID, nil, headers)
}

func (c *Client) CreateTeam(ctx context.Context, body interface{}, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/api/v1/teams", body, headers)
}

func (c *Client) UpdateTeam(ctx context.Context, teamID string, body interface{}, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodPatch, "/api/v1/teams/"+teamID, body, headers)
}

func (c *Client) DeleteTeam(ctx context.Context, teamID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodDelete, "/api/v1/teams/"+teamID, nil, headers)
}

func (c *Client) RunTeam(ctx context.Context, teamID string, body interface{}, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/api/v1/teams/"+teamID+"/run", body, headers)
}

// Tools

func (c *Client) ListTools(ctx context.Context, projectID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodGet, "/api/v1/tools?project_id="+projectID, nil, headers)
}

func (c *Client) GetTool(ctx context.Context, toolID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodGet, "/api/v1/tools/"+toolID, nil, headers)
}

func (c *Client) CreateTool(ctx context.Context, body interface{}, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/api/v1/tools", body, headers)
}

func (c *Client) UpdateTool(ctx context.Context, toolID string, body interface{}, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodPatch, "/api/v1/tools/"+toolID, body, headers)
}

func (c *Client) DeleteTool(ctx context.Context, toolID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodDelete, "/api/v1/tools/"+toolID, nil, headers)
}

// Providers

func (c *Client) ListProviders(ctx context.Context, projectID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodGet, "/api/v1/llm-providers?project_id="+projectID, nil, headers)
}

func (c *Client) GetProvider(ctx context.Context, providerID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodGet, "/api/v1/llm-providers/"+providerID, nil, headers)
}

func (c *Client) CreateProvider(ctx context.Context, body interface{}, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/api/v1/llm-providers", body, headers)
}

func (c *Client) UpdateProvider(ctx context.Context, providerID string, body interface{}, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodPatch, "/api/v1/llm-providers/"+providerID, body, headers)
}

func (c *Client) DeleteProvider(ctx context.Context, providerID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodDelete, "/api/v1/llm-providers/"+providerID, nil, headers)
}

func (c *Client) EnableProvider(ctx context.Context, providerID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/api/v1/llm-providers/"+providerID+"/enable", nil, headers)
}

func (c *Client) DisableProvider(ctx context.Context, providerID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/api/v1/llm-providers/"+providerID+"/disable", nil, headers)
}

func (c *Client) SyncProvider(ctx context.Context, providerID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/api/v1/llm-providers/"+providerID+"/sync", nil, headers)
}

func (c *Client) TestProvider(ctx context.Context, providerID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/api/v1/llm-providers/"+providerID+"/test", nil, headers)
}

// Models

func (c *Client) ListModels(ctx context.Context, projectID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodGet, "/api/v1/models?project_id="+projectID, nil, headers)
}

func (c *Client) GetModel(ctx context.Context, modelID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodGet, "/api/v1/models/"+modelID, nil, headers)
}

// Project AI Configs

func (c *Client) GetProjectAIConfig(ctx context.Context, projectID string, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodGet, "/api/v1/project-ai-configs?project_id="+projectID, nil, headers)
}

func (c *Client) UpsertProjectAIConfig(ctx context.Context, projectID string, body interface{}, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodPut, "/api/v1/project-ai-configs?project_id="+projectID, body, headers)
}

// MCP Project Tools

func (c *Client) ListProjectTools(ctx context.Context, projectID string, params map[string]string) (map[string]interface{}, error) {
	query := "project_id=" + projectID
	for k, v := range params {
		query += "&" + k + "=" + v
	}
	respBody, statusCode, err := c.request(ctx, http.MethodGet, "/api/v1/project-tools?"+query, nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d: %s", statusCode, string(respBody))
	}
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetProjectToolStats(ctx context.Context, projectID string) (map[string]interface{}, error) {
	respBody, statusCode, err := c.request(ctx, http.MethodGet, "/api/v1/project-tools/stats?project_id="+projectID, nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d: %s", statusCode, string(respBody))
	}
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetProjectTool(ctx context.Context, projectID, toolID string) (map[string]interface{}, error) {
	respBody, statusCode, err := c.request(ctx, http.MethodGet, "/api/v1/project-tools/"+toolID+"?project_id="+projectID, nil, nil)
	if err != nil {
		return nil, err
	}
	if statusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d: %s", statusCode, string(respBody))
	}
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) UpdateProjectTool(ctx context.Context, projectID, toolID string, data map[string]interface{}) (map[string]interface{}, error) {
	respBody, statusCode, err := c.request(ctx, http.MethodPut, "/api/v1/project-tools/"+toolID+"?project_id="+projectID, data, nil)
	if err != nil {
		return nil, err
	}
	if statusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d: %s", statusCode, string(respBody))
	}
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) UninstallTool(ctx context.Context, projectID, toolID string) error {
	_, statusCode, err := c.request(ctx, http.MethodDelete, "/api/v1/project-tools/"+toolID+"?project_id="+projectID, nil, nil)
	if err != nil {
		return err
	}
	if statusCode >= 400 {
		return fmt.Errorf("request failed with status %d", statusCode)
	}
	return nil
}

func (c *Client) InstallTool(ctx context.Context, projectID string, data map[string]interface{}) (map[string]interface{}, error) {
	data["project_id"] = projectID
	respBody, statusCode, err := c.request(ctx, http.MethodPost, "/api/v1/project-tools/install", data, nil)
	if err != nil {
		return nil, err
	}
	if statusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d: %s", statusCode, string(respBody))
	}
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) BulkInstallTools(ctx context.Context, projectID string, data map[string]interface{}) ([]map[string]interface{}, error) {
	data["project_id"] = projectID
	respBody, statusCode, err := c.request(ctx, http.MethodPost, "/api/v1/project-tools/bulk-install", data, nil)
	if err != nil {
		return nil, err
	}
	if statusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d: %s", statusCode, string(respBody))
	}
	var result []map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// TeamChat calls the AI team/agent chat endpoint
func (c *Client) TeamChat(ctx context.Context, body map[string]interface{}, headers map[string]string) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/api/v1/agents/run", body, headers)
}

// TeamChatStream calls the AI team/agent chat endpoint with streaming (returns http.Response for SSE)
func (c *Client) TeamChatStream(ctx context.Context, body map[string]interface{}, headers map[string]string) (*http.Response, error) {
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/agents/run", bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Use a client without timeout for streaming
	streamClient := &http.Client{
		Timeout: 0, // No timeout for streaming
	}

	resp, err := streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// GetBaseURL returns the base URL for direct access
func (c *Client) GetBaseURL() string {
	return c.baseURL
}
