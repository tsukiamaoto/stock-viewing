package service

import (
	"fmt"
	"time"

	"stock-viewing-backend/internal/config"
	"stock-viewing-backend/internal/crawler"
	"stock-viewing-backend/internal/database"
	"stock-viewing-backend/internal/model"
)

// ────────────────────────────────────────────────────────────────────
// News Service — orchestrates crawler → translate → database pipeline
// Uses Google Translate (fast) instead of Gemini LLM for translation
// ────────────────────────────────────────────────────────────────────

// FetchMacroNews aggregates CNN, RSS (Reuters/NHK) and Jin10 news.
func FetchMacroNews() []model.NewsItem {
	sections := []crawler.CNNSection{
		{URL: config.Cfg.CNNBusinessURL, Label: "Business"},
		{URL: config.Cfg.CNNWorldURL, Label: "World"},
	}

	// CNN & Reuters: English → Traditional Chinese
	enToZhTW := EnhanceWithTranslate("en")
	cnnNews := crawler.FetchCNNNews(sections, enToZhTW)

	// Reuters & NHK via RSS: auto-detect language → Traditional Chinese
	rssNews := crawler.FetchReutersNews(enToZhTW)
	nhkNews := crawler.FetchNHKNews(EnhanceWithTranslate("ja"))
	allRSS := append(rssNews, nhkNews...)

	// Jin10: Simplified Chinese → Traditional Chinese (with prefix cleaning)
	jin10News := crawler.FetchJin10News(EnhanceJin10)

	// TWSE ETF: Naturally Traditional Chinese
	twseNews := crawler.FetchTWSE_ETFNews(EnhanceWithTranslate("zh-TW"))

	allNews := append(cnnNews, allRSS...)
	allNews = append(allNews, twseNews...)
	return append(allNews, jin10News...)
}

// FetchSymbolNews fetches stock-specific news (placeholder for yfinance equivalent).
func FetchSymbolNews(symbol string) []model.NewsItem {
	if symbol == "Macro" {
		return nil
	}
	// Yahoo Finance news API is not publicly stable — skip for now
	return nil
}

// GetAllNewsForAnalysis aggregates all news and persists to Supabase.
func GetAllNewsForAnalysis(symbol string) []model.NewsItem {
	macro := FetchMacroNews()
	symbolNews := FetchSymbolNews(symbol)
	allNews := append(macro, symbolNews...)
	saveToSupabase(allNews)
	return allNews
}

// ScheduledCrawlTask is called by the background cron job.
func ScheduledCrawlTask() {
	fmt.Printf("[Scheduler] 啟動每分鐘定時爬蟲任務... (%s)\n", time.Now().Format("15:04:05"))
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[Scheduler] 定時爬蟲發生錯誤: %v\n", r)
		}
	}()
	GetAllNewsForAnalysis("Macro")
}

func saveToSupabase(items []model.NewsItem) {
	if len(items) == 0 {
		fmt.Println("[DB] 沒有新聞可以寫入資料庫。")
		return
	}
	fmt.Printf("[DB] 準備寫入 %d 筆新聞至 Supabase...\n", len(items))

	batch := make([]map[string]interface{}, 0, len(items))
	for _, n := range items {
		pubDate := n.PubDate
		if pubDate == "" {
			pubDate = time.Now().UTC().Format(time.RFC3339)
		}
		batch = append(batch, map[string]interface{}{
			"title":              n.Title,
			"translated_title":   n.TranslatedTitle,
			"content":            n.OriginalContent,
			"translated_content": n.Snippet,
			"category":           n.Category,
			"link":               n.Link,
			"source":             n.Source,
			"sourceColor":        n.SourceColor,
			"published_at":       pubDate,
		})
	}

	if err := database.InsertNewsBatch(batch); err != nil {
		fmt.Printf("[DB/Error] 寫入發生錯誤: %v\n", err)
	} else {
		fmt.Println("[DB] 寫入 Supabase 完成。")
	}
}
