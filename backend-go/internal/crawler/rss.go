package crawler

import (
	"stock-viewing-backend/internal/logger"
	"stock-viewing-backend/internal/model"

	"github.com/mmcdole/gofeed"
)

// ────────────────────────────────────────────────────────────────────
// RSS Crawler (Reuters, NHK, generic)
// ────────────────────────────────────────────────────────────────────

// ReutersRSSSources defines Reuters RSS feeds (via Google News proxy).
var ReutersRSSSources = []model.RSSSource{
	{Name: "Reuters Business", URL: "https://news.google.com/rss/search?q=site:reuters.com+business&hl=en-US&gl=US&ceid=US:en", Color: "#ff8000"},
	{Name: "Reuters World", URL: "https://news.google.com/rss/search?q=site:reuters.com+world&hl=en-US&gl=US&ceid=US:en", Color: "#ff8000"},
	{Name: "Reuters Markets", URL: "https://news.google.com/rss/search?q=site:reuters.com+markets&hl=en-US&gl=US&ceid=US:en", Color: "#ff8000"},
}

// NHKRSSSources defines NHK RSS feeds.
var NHKRSSSources = []model.RSSSource{
	{Name: "NHK World", URL: "https://www3.nhk.or.jp/rss/news/cat0.xml", Color: "#0068b7"},
	{Name: "NHK Business", URL: "https://www3.nhk.or.jp/rss/news/cat3.xml", Color: "#0068b7"},
}

// FetchRSSFromSources crawls an arbitrary list of RSS feeds and returns news items.
func FetchRSSFromSources(sources []model.RSSSource, limitPerSource int, enhanceFn func(string, string) model.LLMEnhanceResult) []model.NewsItem {
	parser := gofeed.NewParser()
	var items []model.NewsItem

	for _, src := range sources {
		feed, err := parser.ParseURL(src.URL)
		if err != nil {
			logger.Crawler().Error("RSS fetch failed", "source", src.Name, "error", err)
			logger.RecordFailure(src.Name, 1)
			continue
		}

		count := 0
		for _, entry := range feed.Items {
			if count >= limitPerSource {
				break
			}
			title := CleanText(entry.Title)
			// CleanText now strips HTML tags, fixing Reuters' <a href="..."> issue
			snippet := CleanText(entry.Description)

			llm := enhanceFn(title, snippet)

			translatedTitle := llm.TranslatedTitle
			if translatedTitle == "" {
				translatedTitle = title
			}
			translatedSnippet := llm.TranslatedSnippet
			if translatedSnippet == "" {
				translatedSnippet = snippet
			}
			cat := llm.Category
			if cat == "" {
				cat = "other"
			}

			pubDate := ""
			if entry.PublishedParsed != nil {
				pubDate = entry.PublishedParsed.UTC().Format("Mon, 02 Jan 2006 15:04:05 +0000")
			} else if entry.Published != "" {
				pubDate = entry.Published
			}

			items = append(items, model.NewsItem{
				Title:           title,
				TranslatedTitle: translatedTitle,
				Link:            entry.Link,
				Snippet:         translatedSnippet,
				OriginalContent: snippet,
				Category:        cat,
				PubDate:         pubDate,
				Source:          src.Name,
				SourceColor:     src.Color,
			})
			count++
		}
		if count > 0 {
			logger.RecordSuccess(src.Name, count)
			logger.Crawler().Info("RSS fetch success", "source", src.Name, "count", count)
		}
	}

	// Deduplicate by title
	return deduplicateByTitle(items)
}

// deduplicateByTitle removes duplicate news entries by title.
func deduplicateByTitle(items []model.NewsItem) []model.NewsItem {
	seen := make(map[string]bool)
	unique := make([]model.NewsItem, 0, len(items))
	for _, item := range items {
		if !seen[item.Title] {
			seen[item.Title] = true
			unique = append(unique, item)
		}
	}
	return unique
}

// FetchReutersNews fetches Reuters news from RSS.
func FetchReutersNews(enhanceFn func(string, string) model.LLMEnhanceResult) []model.NewsItem {
	return FetchRSSFromSources(ReutersRSSSources, 10, enhanceFn)
}

// FetchNHKNews fetches NHK news from RSS.
func FetchNHKNews(enhanceFn func(string, string) model.LLMEnhanceResult) []model.NewsItem {
	return FetchRSSFromSources(NHKRSSSources, 10, enhanceFn)
}

// FetchAllRSSNews combines Reuters + NHK.
func FetchAllRSSNews(enhanceFn func(string, string) model.LLMEnhanceResult) []model.NewsItem {
	all := append(ReutersRSSSources, NHKRSSSources...)
	return FetchRSSFromSources(all, 8, enhanceFn)
}
