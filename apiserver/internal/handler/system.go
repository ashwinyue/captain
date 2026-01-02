package handler

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

type SystemHandler struct {
	version string
}

func NewSystemHandler(version string) *SystemHandler {
	return &SystemHandler{version: version}
}

func (h *SystemHandler) GetInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":    "apiserver",
		"version":    h.version,
		"go_version": runtime.Version(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"uptime":     time.Since(startTime).String(),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *SystemHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
