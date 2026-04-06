package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"stock-viewing-backend/internal/crawler"
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
	rg.GET("/twse-etf", getTwseEtfNews)
	rg.GET("/ptt", getPTTNews)
	rg.GET("/cmoney", getCMoneyNews)
}

// GET /api/news/latest
func getLatestNews(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	
	rows, err := database.GetLatestNews(limit, offset)
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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	res := getNewsBySource("CNN", limit, offset)
	res["source"] = "CNN"
	c.JSON(http.StatusOK, res)
}

// GET /api/news/reuters
func getReutersNews(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	res := getNewsBySource("Reuters", limit, offset)
	res["source"] = "Reuters"
	c.JSON(http.StatusOK, res)
}

// GET /api/news/nhk
func getNHKNews(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	res := getNewsBySource("NHK", limit, offset)
	res["source"] = "NHK"
	c.JSON(http.StatusOK, res)
}

// GET /api/news/jin10
func getJin10News(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	res := getNewsBySource("金十", limit, offset)
	res["source"] = "Jin10"
	c.JSON(http.StatusOK, res)
}

// GET /api/news/twse-etf
func getTwseEtfNews(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	res := getNewsBySource("TWSE", limit, offset) // The keyword should match a substring of Source
	res["source"] = "TWSE ETF"
	c.JSON(http.StatusOK, res)
}

// GET /api/news/ptt
func getPTTNews(c *gin.Context) {
	rows, err := database.GetNewsBySource("PTT", 50, 0)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error(), "data": []interface{}{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   mapDBRowsToForumNews(rows),
		"source": "PTT",
	})
}

// GET /api/news/cmoney
func getCMoneyNews(c *gin.Context) {
	symbols := c.DefaultQuery("symbols", "")
	isNew := service.RegisterWatchlistSymbols(symbols)

	// If this is the first time these symbols are requested, fetch them synchronously right now 
	// so the user doesn't see an empty page and have to wait for the next cron interval.
	if isNew {
		items := crawler.FetchCMoneyRealtime(symbols)
		service.SaveForumToSupabase(items)
	}

	rows, err := database.GetNewsBySource("CMoney", 200, 0)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error(), "data": []interface{}{}})
		return
	}

	allForums := mapDBRowsToForumNews(rows)
	
	var filtered []gin.H
	symbolList := strings.Split(symbols, ",")
	symMap := make(map[string]bool)
	for _, s := range symbolList {
		s = strings.TrimSpace(s)
		if s != "" {
			symMap[s] = true
		}
	}
	
	for _, f := range allForums {
		sym, ok := f["symbol"].(string)
		if ok && (len(symMap) == 0 || symMap[sym]) {
			filtered = append(filtered, f)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   filtered,
		"source": "CMoney股市爆料同學會",
	})
}

// ────────────────────────────────────────────────────────────────────
// Internal helpers
// ────────────────────────────────────────────────────────────────────

func getNewsBySource(keyword string, limit int, offset int) gin.H {
	rows, err := database.GetNewsBySource(keyword, limit, offset)
	if err != nil {
		return gin.H{"status": "error", "message": err.Error(), "data": []interface{}{}}
	}
	return gin.H{"status": "success", "data": mapDBRowsToNews(rows)}
}

func mapDBRowsToForumNews(rows []map[string]interface{}) []gin.H {
	result := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		commentsStr := getStr(row, "content")
		var comments []map[string]string
		if len(commentsStr) > 0 {
			_ = json.Unmarshal([]byte(commentsStr), &comments)
		}
		
		titleStr := getStr(row, "title")
		symbolStr := ""
		if strings.HasPrefix(titleStr, "[") {
			parts := strings.SplitN(titleStr, "]", 2)
			if len(parts) == 2 {
				symbolStr = strings.TrimPrefix(parts[0], "[")
			}
		}

		result = append(result, gin.H{
			"title":            titleStr,
			"translated_title": getStr(row, "translated_title"),
			"snippet":          getStr(row, "translated_content"),
			"original_content": getStr(row, "translated_content"),
			"category":         getStr(row, "category"),
			"link":             getStr(row, "link"),
			"source":           getStr(row, "source"),
			"sourceColor":      getStr(row, "sourceColor"),
			"pubDate":          getStr(row, "published_at"),
			"comments":         comments,
			"symbol":           symbolStr,
		})
	}
	return result
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
