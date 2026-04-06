package handler

import (
	"bufio"
	"net/http"
	"os"

	"stock-viewing-backend/internal/logger"

	"github.com/gin-gonic/gin"
)

// RegisterSystemRoutes registers diagnostic / system routes.
func RegisterSystemRoutes(rg *gin.RouterGroup) {
	rg.GET("/logs", getLogs)
	rg.GET("/stats", getCrawlerStats)
}

func getLogs(c *gin.Context) {
	// Simple tail for the recent logs (max 500 lines for web viewer)
	path := "logs/crawler.log"
	file, err := os.Open(path)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "No logs found yet"})
		return
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Tail last 500 lines
	maxLines := 500
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   lines,
	})
}

func getCrawlerStats(c *gin.Context) {
	stats := logger.GetStats()
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}
