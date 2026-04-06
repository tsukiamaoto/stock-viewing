package handler

import (
	"net/http"

	"stock-viewing-backend/internal/database"
	"stock-viewing-backend/internal/llm"
	"stock-viewing-backend/internal/model"
	"stock-viewing-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterNewsRoutes registers all /api/news/* endpoints.
func RegisterNewsRoutes(rg *gin.RouterGroup) {
	rg.GET("/latest", getLatestNews)
	rg.GET("/categorize/:symbol", categorizeNews)
	rg.GET("/cnn", getCNNNews)
	rg.GET("/reuters", getReutersNews)
	rg.GET("/nhk", getNHKNews)
	rg.GET("/jin10", getJin10News)
}

// GET /api/news/latest
func getLatestNews(c *gin.Context) {
	rows, err := database.GetLatestNews(50)
	if err != nil {
		c.JSON(http.StatusOK, model.NewError(err.Error()))
		return
	}
	data := mapDBRowsToNews(rows)
	c.JSON(http.StatusOK, model.NewSuccess(data))
}

// GET /api/news/categorize/:symbol
func categorizeNews(c *gin.Context) {
	symbol := c.Param("symbol")
	rawNews := service.GetAllNewsForAnalysis(symbol)

	limit := 20
	if len(rawNews) < limit {
		limit = len(rawNews)
	}
	categorized := llm.CategorizeNewsWithLLM(rawNews[:limit])

	c.JSON(http.StatusOK, gin.H{
		"symbol": symbol,
		"status": "success",
		"data":   categorized,
	})
}

// GET /api/news/cnn
func getCNNNews(c *gin.Context) {
	res := getNewsBySource("CNN", 15)
	res["source"] = "CNN"
	c.JSON(http.StatusOK, res)
}

// GET /api/news/reuters
func getReutersNews(c *gin.Context) {
	res := getNewsBySource("Reuters", 15)
	res["source"] = "Reuters"
	c.JSON(http.StatusOK, res)
}

// GET /api/news/nhk
func getNHKNews(c *gin.Context) {
	res := getNewsBySource("NHK", 15)
	res["source"] = "NHK"
	c.JSON(http.StatusOK, res)
}

// GET /api/news/jin10
func getJin10News(c *gin.Context) {
	res := getNewsBySource("金十", 15)
	res["source"] = "Jin10"
	c.JSON(http.StatusOK, res)
}

// ────────────────────────────────────────────────────────────────────
// Internal helpers
// ────────────────────────────────────────────────────────────────────

func getNewsBySource(keyword string, limit int) gin.H {
	rows, err := database.GetNewsBySource(keyword, limit)
	if err != nil {
		return gin.H{"status": "error", "message": err.Error(), "data": []interface{}{}}
	}
	return gin.H{"status": "success", "data": mapDBRowsToNews(rows)}
}

func mapDBRowsToNews(rows []map[string]interface{}) []gin.H {
	result := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		snippet := getStr(row, "translated_content")
		if snippet == "" {
			snippet = getStr(row, "content")
		}
		result = append(result, gin.H{
			"title":            getStr(row, "title"),
			"translated_title": getStr(row, "translated_title"),
			"snippet":          snippet,
			"original_snippet": getStr(row, "content"),
			"category":         getStr(row, "category"),
			"link":             getStr(row, "link"),
			"source":           getStr(row, "source"),
			"sourceColor":      getStr(row, "sourceColor"),
			"pubDate":          getStr(row, "published_at"),
		})
	}
	return result
}

func getStr(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
