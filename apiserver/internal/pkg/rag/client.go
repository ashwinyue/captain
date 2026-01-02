package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
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
			Timeout: 120 * time.Second,
		},
	}
}

func (c *Client) request(ctx context.Context, method, path string, body interface{}, params map[string]string) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	// Build URL with query params
	reqURL := c.baseURL + path
	if len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Set(k, v)
		}
		reqURL += "?" + values.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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

func (c *Client) uploadFile(ctx context.Context, path string, params map[string]string, filename string, fileContent io.Reader, contentType string) ([]byte, int, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, 0, err
	}
	if _, err := io.Copy(part, fileContent); err != nil {
		return nil, 0, err
	}

	// Add other params as form fields
	for k, v := range params {
		writer.WriteField(k, v)
	}

	writer.Close()

	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, body)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return respBody, resp.StatusCode, nil
}

// Collections

func (c *Client) ListCollections(ctx context.Context, projectID string, params map[string]string) ([]byte, int, error) {
	if params == nil {
		params = make(map[string]string)
	}
	params["project_id"] = projectID
	return c.request(ctx, http.MethodGet, "/v1/collections", nil, params)
}

func (c *Client) CreateCollection(ctx context.Context, projectID string, body interface{}) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/v1/collections", body, map[string]string{"project_id": projectID})
}

func (c *Client) GetCollection(ctx context.Context, projectID, collectionID string, includeStats bool) ([]byte, int, error) {
	params := map[string]string{"project_id": projectID}
	if includeStats {
		params["include_stats"] = "true"
	}
	return c.request(ctx, http.MethodGet, "/v1/collections/"+collectionID, nil, params)
}

func (c *Client) UpdateCollection(ctx context.Context, projectID, collectionID string, body interface{}) ([]byte, int, error) {
	return c.request(ctx, http.MethodPut, "/v1/collections/"+collectionID, body, map[string]string{"project_id": projectID})
}

func (c *Client) DeleteCollection(ctx context.Context, projectID, collectionID string) ([]byte, int, error) {
	return c.request(ctx, http.MethodDelete, "/v1/collections/"+collectionID, nil, map[string]string{"project_id": projectID})
}

func (c *Client) SearchCollectionDocuments(ctx context.Context, projectID, collectionID string, body interface{}) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/v1/collections/"+collectionID+"/documents/search", body, map[string]string{"project_id": projectID})
}

// Files

func (c *Client) ListFiles(ctx context.Context, projectID string, params map[string]string) ([]byte, int, error) {
	if params == nil {
		params = make(map[string]string)
	}
	params["project_id"] = projectID
	return c.request(ctx, http.MethodGet, "/v1/files", nil, params)
}

func (c *Client) UploadFile(ctx context.Context, projectID string, filename string, fileContent io.Reader, contentType string, params map[string]string) ([]byte, int, error) {
	if params == nil {
		params = make(map[string]string)
	}
	params["project_id"] = projectID
	return c.uploadFile(ctx, "/v1/files", params, filename, fileContent, contentType)
}

func (c *Client) GetFile(ctx context.Context, projectID, fileID string) ([]byte, int, error) {
	return c.request(ctx, http.MethodGet, "/v1/files/"+fileID, nil, map[string]string{"project_id": projectID})
}

func (c *Client) DeleteFile(ctx context.Context, projectID, fileID string) ([]byte, int, error) {
	return c.request(ctx, http.MethodDelete, "/v1/files/"+fileID, nil, map[string]string{"project_id": projectID})
}

func (c *Client) DownloadFile(ctx context.Context, projectID, fileID string) (*http.Response, error) {
	reqURL := c.baseURL + "/v1/files/" + fileID + "/download?project_id=" + projectID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	return c.httpClient.Do(req)
}

// Website Pages

func (c *Client) ListWebsitePages(ctx context.Context, projectID, collectionID string, params map[string]string) ([]byte, int, error) {
	if params == nil {
		params = make(map[string]string)
	}
	params["project_id"] = projectID
	params["collection_id"] = collectionID
	return c.request(ctx, http.MethodGet, "/v1/websites/pages", nil, params)
}

func (c *Client) GetWebsitePage(ctx context.Context, projectID, pageID string) ([]byte, int, error) {
	return c.request(ctx, http.MethodGet, "/v1/websites/pages/"+pageID, nil, map[string]string{"project_id": projectID})
}

func (c *Client) AddWebsitePage(ctx context.Context, projectID, collectionID string, body interface{}) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/v1/websites/pages", body, map[string]string{"project_id": projectID, "collection_id": collectionID})
}

func (c *Client) DeleteWebsitePage(ctx context.Context, projectID, pageID string) ([]byte, int, error) {
	return c.request(ctx, http.MethodDelete, "/v1/websites/pages/"+pageID, nil, map[string]string{"project_id": projectID})
}

func (c *Client) RecrawlWebsitePage(ctx context.Context, projectID, pageID string) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/v1/websites/pages/"+pageID+"/recrawl", nil, map[string]string{"project_id": projectID})
}

func (c *Client) CrawlDeeperFromPage(ctx context.Context, projectID, pageID string, body interface{}) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/v1/websites/pages/"+pageID+"/crawl-deeper", body, map[string]string{"project_id": projectID})
}

func (c *Client) GetCrawlProgress(ctx context.Context, projectID, collectionID string) ([]byte, int, error) {
	return c.request(ctx, http.MethodGet, "/v1/websites/progress", nil, map[string]string{"project_id": projectID, "collection_id": collectionID})
}

// QA Pairs

func (c *Client) ListQAPairs(ctx context.Context, projectID, collectionID string, params map[string]string) ([]byte, int, error) {
	if params == nil {
		params = make(map[string]string)
	}
	params["project_id"] = projectID
	return c.request(ctx, http.MethodGet, "/v1/collections/"+collectionID+"/qa-pairs", nil, params)
}

func (c *Client) CreateQAPair(ctx context.Context, projectID, collectionID string, body interface{}) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/v1/collections/"+collectionID+"/qa-pairs", body, map[string]string{"project_id": projectID})
}

func (c *Client) BatchCreateQAPairs(ctx context.Context, projectID, collectionID string, body interface{}) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/v1/collections/"+collectionID+"/qa-pairs/batch", body, map[string]string{"project_id": projectID})
}

func (c *Client) ImportQAPairs(ctx context.Context, projectID, collectionID string, body interface{}) ([]byte, int, error) {
	return c.request(ctx, http.MethodPost, "/v1/collections/"+collectionID+"/qa-pairs/import", body, map[string]string{"project_id": projectID})
}

func (c *Client) GetQAPair(ctx context.Context, projectID, qaPairID string) ([]byte, int, error) {
	return c.request(ctx, http.MethodGet, "/v1/qa-pairs/"+qaPairID, nil, map[string]string{"project_id": projectID})
}

func (c *Client) UpdateQAPair(ctx context.Context, projectID, qaPairID string, body interface{}) ([]byte, int, error) {
	return c.request(ctx, http.MethodPut, "/v1/qa-pairs/"+qaPairID, body, map[string]string{"project_id": projectID})
}

func (c *Client) DeleteQAPair(ctx context.Context, projectID, qaPairID string) ([]byte, int, error) {
	return c.request(ctx, http.MethodDelete, "/v1/qa-pairs/"+qaPairID, nil, map[string]string{"project_id": projectID})
}

func (c *Client) ListQACategories(ctx context.Context, projectID string, collectionID string) ([]byte, int, error) {
	params := map[string]string{"project_id": projectID}
	if collectionID != "" {
		params["collection_id"] = collectionID
	}
	return c.request(ctx, http.MethodGet, "/v1/qa-categories", nil, params)
}

func (c *Client) GetQAStats(ctx context.Context, projectID, collectionID string) ([]byte, int, error) {
	return c.request(ctx, http.MethodGet, "/v1/collections/"+collectionID+"/qa-pairs/stats", nil, map[string]string{"project_id": projectID})
}
