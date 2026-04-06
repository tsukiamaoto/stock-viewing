package handler

import (
	"net/http"

	"stock-viewing-backend/internal/model"
	"stock-viewing-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterStockDetailRoutes registers the stock detail endpoint.
func RegisterStockDetailRoutes(rg *gin.RouterGroup) {
	rg.GET("/detail/:code", getStockDetail)
}

// GET /api/stocks/detail/:code
func getStockDetail(c *gin.Context) {
	code := c.Param("code")

	detail, err := service.GetStockDetail(code)
	if err != nil {
		c.JSON(http.StatusOK, model.NewError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.NewSuccess(detail))
}
