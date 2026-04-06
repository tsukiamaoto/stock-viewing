package main

import (
	"fmt"
	"stock-viewing-backend/internal/config"
	"stock-viewing-backend/internal/service"
)

func main() {
	_, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}
	fmt.Println("Starting manual crawl...")
	service.ScheduledCrawlTask()
	fmt.Println("Manual crawl finished!")
}
