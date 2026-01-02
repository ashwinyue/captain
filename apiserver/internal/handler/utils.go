package handler

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/html"
)

const (
	requestTimeout   = 10 * time.Second
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

type UtilsHandler struct{}

func NewUtilsHandler() *UtilsHandler {
	return &UtilsHandler{}
}

// WebsiteMetadataRequest represents the request for extracting website metadata
type WebsiteMetadataRequest struct {
	URL string `json:"url" binding:"required,url"`
}

// WebsiteMetadataResponse represents the response for website metadata extraction
type WebsiteMetadataResponse struct {
	URL         string  `json:"url"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Favicon     *string `json:"favicon,omitempty"`
	OGImage     *string `json:"og_image,omitempty"`
	Success     bool    `json:"success"`
	Error       *string `json:"error,omitempty"`
}

// ExtractWebsiteMetadata extracts metadata from a website URL
func (h *UtilsHandler) ExtractWebsiteMetadata(c *gin.Context) {
	var req WebsiteMetadataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL format"})
		return
	}

	client := &http.Client{
		Timeout: requestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, req.URL, nil)
	if err != nil {
		errStr := err.Error()
		c.JSON(http.StatusOK, WebsiteMetadataResponse{
			URL:     req.URL,
			Success: false,
			Error:   &errStr,
		})
		return
	}
	httpReq.Header.Set("User-Agent", defaultUserAgent)

	resp, err := client.Do(httpReq)
	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			c.JSON(http.StatusGatewayTimeout, gin.H{"detail": "Request timeout while fetching " + req.URL})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"detail": "Failed to connect to " + req.URL})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		c.JSON(http.StatusBadGateway, gin.H{"detail": "Website returned status " + resp.Status})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errStr := err.Error()
		c.JSON(http.StatusOK, WebsiteMetadataResponse{
			URL:     req.URL,
			Success: false,
			Error:   &errStr,
		})
		return
	}

	metadata := extractMetadata(string(body), req.URL)
	c.JSON(http.StatusOK, WebsiteMetadataResponse{
		URL:         req.URL,
		Title:       metadata.Title,
		Description: metadata.Description,
		Favicon:     metadata.Favicon,
		OGImage:     metadata.OGImage,
		Success:     true,
	})
}

type metadata struct {
	Title       *string
	Description *string
	Favicon     *string
	OGImage     *string
}

func extractMetadata(htmlContent, baseURL string) metadata {
	var m metadata

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return m
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
					title := strings.TrimSpace(n.FirstChild.Data)
					m.Title = &title
				}
			case "meta":
				var name, property, content string
				for _, attr := range n.Attr {
					switch attr.Key {
					case "name":
						name = strings.ToLower(attr.Val)
					case "property":
						property = strings.ToLower(attr.Val)
					case "content":
						content = attr.Val
					}
				}
				if name == "description" && content != "" && m.Description == nil {
					desc := strings.TrimSpace(content)
					m.Description = &desc
				}
				if property == "og:description" && content != "" && m.Description == nil {
					desc := strings.TrimSpace(content)
					m.Description = &desc
				}
				if property == "og:image" && content != "" {
					m.OGImage = &content
				}
			case "link":
				var rel, href string
				for _, attr := range n.Attr {
					switch attr.Key {
					case "rel":
						rel = strings.ToLower(attr.Val)
					case "href":
						href = attr.Val
					}
				}
				if strings.Contains(rel, "icon") && href != "" && m.Favicon == nil {
					favicon := makeAbsoluteURL(href, baseURL)
					m.Favicon = &favicon
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return m
}

func makeAbsoluteURL(href, baseURL string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	if strings.HasPrefix(href, "//") {
		return "https:" + href
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return href
	}

	if strings.HasPrefix(href, "/") {
		return base.Scheme + "://" + base.Host + href
	}

	// Remove query and fragment from base
	base.RawQuery = ""
	base.Fragment = ""

	// Handle relative paths
	basePath := base.Path
	if !strings.HasSuffix(basePath, "/") {
		// Remove the last path segment
		if idx := strings.LastIndex(basePath, "/"); idx >= 0 {
			basePath = basePath[:idx+1]
		}
	}

	return base.Scheme + "://" + base.Host + basePath + href
}
