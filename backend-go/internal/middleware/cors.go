package middleware

import (
	"stock-viewing-backend/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns a gin middleware configured with the allowed origins from config.
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     config.Cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	})
}
