package service

import (
	"encoding/json"
	"strings"
	"time"

	"stock-viewing-backend/internal/config"
	"stock-viewing-backend/internal/crawler"
	"stock-viewing-backend/internal/database"
	"stock-viewing-backend/internal/logger"
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
	logger.Crawler().Info("Scheduled crawl started", "time", time.Now().Format("15:04:05"))
	defer func() {
		if r := recover(); r != nil {
			logger.Crawler().Error("Scheduled crawl panic", "panic", r)
		}
	}()
	// Fetch general macro news
	GetAllNewsForAnalysis("Macro")
	
	// Fetch PTT forum
	pttItems := crawler.FetchPTTStockRealtime()
	SaveForumToSupabase(pttItems)

	// Fetch tracked CMoney forums
	symbols := GetActiveSymbols()
	if len(symbols) > 0 {
		logger.Crawler().Info("Background crawl active CMoney symbols", "symbols", symbols)
		cmoneyItems := crawler.FetchCMoneyRealtime(strings.Join(symbols, ","))
		SaveForumToSupabase(cmoneyItems)
	}
}

func SaveForumToSupabase(items []map[string]interface{}) {
	if len(items) == 0 {
		return
	}
	batch := make([]map[string]interface{}, 0, len(items))
	for _, n := range items {
		// Serialize comments array to JSON string to fit in the 'content' column
		var commentsStr = "[]"
		if comments, ok := n["comments"]; ok {
			if b, err := json.Marshal(comments); err == nil {
				commentsStr = string(b)
			}
		}

		batch = append(batch, map[string]interface{}{
			"title":              n["title"],
			"translated_title":   n["title"],     // No translation for forum
			"content":            commentsStr,    // Serialized comments!
			"translated_content": n["snippet"],   // Snippet
			"category":           n["category"],  // Author
			"link":               n["link"],
			"source":             n["source"],
			"sourceColor":        n["sourceColor"],
			"published_at":       n["pubDate"],
		})
	}
	if err := database.InsertNewsBatch(batch); err != nil {
		logger.Crawler().Error("Forum DB Write failed", "error", err)
	} else {
		logger.Crawler().Info("Forum DB Write success", "count", len(batch))
	}
}

func saveToSupabase(items []model.NewsItem) {
	if len(items) == 0 {
		logger.Crawler().Info("DB: No news to write")
		return
	}
	logger.Crawler().Info("DB: Writing news batch", "count", len(items))

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
		logger.Crawler().Error("DB Write failed", "error", err)
	} else {
		logger.Crawler().Info("DB Write success")
	}
}
