package apiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Client is a client for the apiserver internal API
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new apiserver client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// AIServiceEvent represents an event to send to the apiserver
type AIServiceEvent struct {
	EventType string                 `json:"event_type"`
	VisitorID *uuid.UUID             `json:"visitor_id,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

// AIEventResponse represents the response from the AI events endpoint
type AIEventResponse struct {
	EventType string                 `json:"event_type"`
	Result    map[string]interface{} `json:"result"`
}

// SendAIEvent sends an AI event to the apiserver internal endpoint
func (c *Client) SendAIEvent(ctx context.Context, event *AIServiceEvent) (*AIEventResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("apiserver internal URL not configured")
	}

	jsonBody, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("marshal event: %w", err)
	}

	url := c.baseURL + "/v1/internal/ai-events"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("apiserver returned %d: %s", resp.StatusCode, string(body))
	}

	var result AIEventResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &result, nil
}

// SendManualServiceRequest sends a manual service request event
func (c *Client) SendManualServiceRequest(ctx context.Context, visitorID uuid.UUID, reason string) (*AIEventResponse, error) {
	event := &AIServiceEvent{
		EventType: "manual_service.request",
		VisitorID: &visitorID,
		Payload: map[string]interface{}{
			"reason": reason,
		},
	}
	return c.SendAIEvent(ctx, event)
}

// SendVisitorInfoUpdate sends a visitor info update event
func (c *Client) SendVisitorInfoUpdate(ctx context.Context, visitorID uuid.UUID, info map[string]interface{}) (*AIEventResponse, error) {
	event := &AIServiceEvent{
		EventType: "visitor_info.update",
		VisitorID: &visitorID,
		Payload:   info,
	}
	return c.SendAIEvent(ctx, event)
}

// SendVisitorTagAdd sends a visitor tag add event
func (c *Client) SendVisitorTagAdd(ctx context.Context, visitorID uuid.UUID, tags []map[string]string) (*AIEventResponse, error) {
	tagsInterface := make([]interface{}, len(tags))
	for i, t := range tags {
		tagsInterface[i] = t
	}
	event := &AIServiceEvent{
		EventType: "visitor_tag.add",
		VisitorID: &visitorID,
		Payload: map[string]interface{}{
			"tags": tagsInterface,
		},
	}
	return c.SendAIEvent(ctx, event)
}

// SendVisitorSentimentUpdate sends a visitor sentiment update event
func (c *Client) SendVisitorSentimentUpdate(ctx context.Context, visitorID uuid.UUID, sentiment map[string]interface{}) (*AIEventResponse, error) {
	event := &AIServiceEvent{
		EventType: "visitor_sentiment.update",
		VisitorID: &visitorID,
		Payload: map[string]interface{}{
			"sentiment": sentiment,
		},
	}
	return c.SendAIEvent(ctx, event)
}

// GetVisitorInfo gets visitor information from apiserver
func (c *Client) GetVisitorInfo(ctx context.Context, projectID, visitorID string) (map[string]interface{}, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("apiserver internal URL not configured")
	}

	url := fmt.Sprintf("%s/internal/visitors/%s?project_id=%s", c.baseURL, visitorID, projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("visitor not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("apiserver returned %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return result, nil
}
