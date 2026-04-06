package handler

import (
	"net/http"

	"stock-viewing-backend/internal/model"
	"stock-viewing-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterShareholdersRoutes registers shareholder distribution endpoints.
func RegisterShareholdersRoutes(rg *gin.RouterGroup) {
	rg.GET("/shareholders/:code", getShareholderDistribution)
}

// GET /api/stocks/shareholders/:code
func getShareholderDistribution(c *gin.Context) {
	code := c.Param("code")

	data, err := service.GetShareholderDistribution(code)
	if err != nil {
		c.JSON(http.StatusOK, model.NewError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.NewSuccess(data))
}
