package handler

import (
	"net/http"
	"strings"

	"stock-viewing-backend/internal/model"
	"stock-viewing-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterQuotesRoutes registers stock quote endpoints.
func RegisterQuotesRoutes(rg *gin.RouterGroup) {
	rg.GET("/index", getIndexQuote)
	rg.GET("/watchlist", getWatchlistQuotes)
}

// GET /api/stocks/index?yf_symbol=^KS11
func getIndexQuote(c *gin.Context) {
	yfSymbol := c.Query("yf_symbol")
	if yfSymbol == "" {
		c.JSON(http.StatusBadRequest, model.NewError("yf_symbol is required"))
		return
	}

	quote, err := service.GetIndexQuote(yfSymbol)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, model.NewSuccess(quote))
}

// GET /api/stocks/watchlist?symbols=2330,2317
func getWatchlistQuotes(c *gin.Context) {
	symbolsStr := c.Query("symbols")
	if symbolsStr == "" {
		c.JSON(http.StatusBadRequest, model.NewError("No symbols provided"))
		return
	}

	var symList []string
	for _, s := range strings.Split(symbolsStr, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			symList = append(symList, s)
		}
	}
	if len(symList) == 0 {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": "No symbols provided", "data": []interface{}{}})
		return
	}

	results := service.GetWatchlistQuotes(symList)
	c.JSON(http.StatusOK, model.NewSuccess(results))
}
