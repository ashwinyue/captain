package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type DocsHandler struct {
	aiServiceURL       string
	ragServiceURL      string
	platformServiceURL string
}

func NewDocsHandler(aiServiceURL, ragServiceURL, platformServiceURL string) *DocsHandler {
	return &DocsHandler{
		aiServiceURL:       aiServiceURL,
		ragServiceURL:      ragServiceURL,
		platformServiceURL: platformServiceURL,
	}
}

var services = map[string]string{
	"api":      "TGO API Service",
	"ai":       "AI Service",
	"rag":      "RAG Service",
	"platform": "Platform Service",
}

func (h *DocsHandler) getServiceURL(service string) string {
	switch service {
	case "ai":
		return h.aiServiceURL
	case "rag":
		return h.ragServiceURL
	case "platform":
		return h.platformServiceURL
	default:
		return ""
	}
}

// Index lists all available service documentation
func (h *DocsHandler) Index(c *gin.Context) {
	links := ""
	for key, name := range services {
		links += fmt.Sprintf(`<li><a href="/api/v1/docs/%s">%s</a></li>`, key, name)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><title>API Documentation</title>
<style>
    body { font-family: Arial, sans-serif; padding: 40px; background: #f5f5f5; }
    .container { background: white; padding: 30px; border-radius: 8px; max-width: 600px; margin: 0 auto; }
    h1 { color: #333; } ul { line-height: 2; } a { color: #3498db; }
</style>
</head>
<body><div class="container"><h1>📚 API Documentation</h1><ul>%s</ul></div></body>
</html>`, links)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// ServiceDocs shows Swagger UI for a service
func (h *DocsHandler) ServiceDocs(c *gin.Context) {
	service := c.Param("service")
	name, ok := services[service]
	if !ok {
		h.errorHTML(c, "Not Found", fmt.Sprintf("Service '%s' not found.", service))
		return
	}

	var openapiURL string
	if service == "api" {
		openapiURL = "/api/v1/openapi.json"
	} else {
		openapiURL = fmt.Sprintf("/api/v1/docs/%s/openapi.json", service)
	}

	h.swaggerUI(c, name, openapiURL)
}

// ProxyOpenAPI proxies OpenAPI JSON from remote service
func (h *DocsHandler) ProxyOpenAPI(c *gin.Context) {
	service := c.Param("service")
	if service == "api" {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Use /api/v1/openapi.json for API service"})
		return
	}

	serviceURL := h.getServiceURL(service)
	if serviceURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Service not found"})
		return
	}

	url := serviceURL + "/openapi.json"
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"detail": "Service unavailable"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"detail": "Failed to read response"})
		return
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"detail": "Invalid OpenAPI JSON"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *DocsHandler) swaggerUI(c *gin.Context, title, openapiURL string) {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>%s - Swagger UI</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({
            url: "%s",
            dom_id: '#swagger-ui',
            presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
            layout: "StandaloneLayout"
        });
    </script>
</body>
</html>`, title, openapiURL)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

func (h *DocsHandler) errorHTML(c *gin.Context, title, message string) {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>%s - Error</title>
    <style>
        body { font-family: Arial, sans-serif; padding: 40px; background: #f5f5f5; }
        .error { background: white; padding: 30px; border-radius: 8px; max-width: 600px; margin: 0 auto; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #e74c3c; }
        p { color: #666; }
    </style>
</head>
<body>
    <div class="error">
        <h1>⚠️ %s</h1>
        <p>%s</p>
    </div>
</body>
</html>`, title, title, message)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusNotFound, html)
}
