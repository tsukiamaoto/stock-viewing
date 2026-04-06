package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"stock-viewing-backend/internal/config"
	"stock-viewing-backend/internal/handler"
	"stock-viewing-backend/internal/middleware"
	"stock-viewing-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

func main() {
	// ── 1. Load configuration ────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// ── 2. Setup Gin router ──────────────────────────────────────
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(middleware.CORS())

	// ── 3. Register routes (same structure as Python FastAPI) ─────
	stocksGroup := r.Group("/api/stocks")
	{
		handler.RegisterQuotesRoutes(stocksGroup)
		handler.RegisterStockDetailRoutes(stocksGroup)
		handler.RegisterShareholdersRoutes(stocksGroup)
	}

	newsGroup := r.Group("/api/news")
	{
		handler.RegisterNewsRoutes(newsGroup)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().UTC().Format(time.RFC3339)})
	})

	// ── 4. Setup background scheduler (replaces APScheduler) ─────
	scheduler := cron.New()
	cronSpec := fmt.Sprintf("@every %dm", cfg.CrawlerIntervalMinutes)
	_, err = scheduler.AddFunc(cronSpec, service.ScheduledCrawlTask)
	if err != nil {
		log.Printf("[Scheduler] 排程器設定失敗: %v", err)
	} else {
		scheduler.Start()
		fmt.Printf("[Scheduler] 排程器已啟動，設定為每 %d 分鐘執行一次爬取。\n", cfg.CrawlerIntervalMinutes)
	}

	// ── 5. Start HTTP server with graceful shutdown ──────────────
	srv := &http.Server{
		Addr:    cfg.Addr(),
		Handler: r,
	}

	go func() {
		fmt.Printf("啟動 Go 後端伺服器在 %s...\n", cfg.Addr())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("\n正在優雅關閉伺服器...")

	scheduler.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced shutdown: %v", err)
	}
	fmt.Println("伺服器已關閉。")
}
