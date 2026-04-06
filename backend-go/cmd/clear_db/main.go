package main

import (
	"fmt"
	"net/http"
	"stock-viewing-backend/internal/config"
	"strings"
)

func main() {
	config.LoadConfig()
	if config.Cfg.SupabaseURL == "" {
		fmt.Println("No Supabase connection configured.")
		return
	}

	// Delete all rows in the news table by passing a condition that matches everything, 
	// e.g. title is not null or id > 0
	url := strings.TrimRight(config.Cfg.SupabaseURL, "/") + "/rest/v1/news?link=not.is.null"
	
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("apikey", config.Cfg.SupabaseKey)
	req.Header.Set("Authorization", "Bearer "+config.Cfg.SupabaseKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println("Successfully cleared the news table in Supabase.")
	} else {
		fmt.Printf("Failed to clear news table. Status: %d\n", resp.StatusCode)
	}
}
