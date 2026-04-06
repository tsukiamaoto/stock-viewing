package crawler

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"stock-viewing-backend/internal/model"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

// ────────────────────────────────────────────────────────────────────
// Jin10 (金十數據) Flash News Crawler
// Three-tier strategy: Flash API → RSSHub fallback → direct scrape
// ────────────────────────────────────────────────────────────────────

// FetchJin10News attempts to get Jin10 flash news from multiple sources.
func FetchJin10News(enhanceFn func(string, string) model.LLMEnhanceResult) []model.NewsItem {
	items := fetchJin10FlashAPI(enhanceFn)
	if len(items) > 0 {
		fmt.Printf("[Jin10] 共取得 %d 則快訊 (Flash API)\n", len(items))
		return deduplicateNews(items)
	}

	items = fetchJin10RSSHub(enhanceFn)
	if len(items) > 0 {
		fmt.Printf("[Jin10] 共取得 %d 則快訊 (RSSHub)\n", len(items))
		return deduplicateNews(items)
	}

	items = fetchJin10DirectScrape(enhanceFn)
	fmt.Printf("[Jin10] 共取得 %d 則快訊 (Direct Scrape)\n", len(items))
	return deduplicateNews(items)
}

func fetchJin10FlashAPI(enhanceFn func(string, string) model.LLMEnhanceResult) []model.NewsItem {
	extra := map[string]string{
		"Referer": "https://www.jin10.com/",
		"Origin":  "https://www.jin10.com",
	}
	body, err := FetchURL("https://flash-api.jin10.com/get_flash?channel=-9999&vip=1", extra)
	if err != nil {
		fmt.Printf("[Jin10] Flash API 請求失敗: %v\n", err)
		return nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		fmt.Printf("[Jin10] Flash API 解析失敗: %v\n", err)
		return nil
	}

	var flashList []interface{}
	if d, ok := raw["data"]; ok {
		switch v := d.(type) {
		case map[string]interface{}:
			if arr, ok := v["data"].([]interface{}); ok {
				flashList = arr
			}
		case []interface{}:
			flashList = v
		}
	}

	var items []model.NewsItem
	for i, item := range flashList {
		if i >= 20 {
			break
		}
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		content := extractJin10Content(m)
		// Use CleanJin10Text to strip "分享收藏详情复制", timestamps, VIP tags
		content = CleanJin10Text(content)
		if len(content) < 5 {
			continue
		}

		pubTime := ""
		if t, ok := m["time"].(string); ok {
			pubTime = t
		}
		if pubTime == "" {
			pubTime = time.Now().UTC().Format(time.RFC3339)
		}

		titlePart := content
		if utf8.RuneCountInString(titlePart) > 80 {
			runes := []rune(titlePart)
			titlePart = string(runes[:80])
		}

		llm := enhanceFn(titlePart, content)

		translatedTitle := llm.TranslatedTitle
		if translatedTitle == "" {
			translatedTitle = titlePart
		}
		snippet := llm.TranslatedSnippet
		if snippet == "" {
			snippet = content
		}
		cat := llm.Category
		if cat == "" {
			cat = "other"
		}

		items = append(items, model.NewsItem{
			Title:           titlePart,
			TranslatedTitle: translatedTitle,
			Link:            "https://www.jin10.com/",
			Snippet:         snippet,
			OriginalContent: content,
			PubDate:         pubTime,
			Source:          "金十數據",
			SourceColor:     "#c8a000",
			Category:        cat,
		})
	}
	return items
}

func extractJin10Content(m map[string]interface{}) string {
	if d, ok := m["data"].(map[string]interface{}); ok {
		if c, ok := d["content"].(string); ok {
			return c
		}
	}
	if c, ok := m["content"].(string); ok {
		return c
	}
	return ""
}

func fetchJin10RSSHub(enhanceFn func(string, string) model.LLMEnhanceResult) []model.NewsItem {
	fmt.Println("[Jin10] 改用 RSSHub 代理抓取...")
	urls := []string{
		"https://rsshub.app/jin10",
		"https://rss.fatcat.app/jin10",
	}
	parser := gofeed.NewParser()

	for _, rssURL := range urls {
		feed, err := parser.ParseURL(rssURL)
		if err != nil {
			fmt.Printf("[Jin10] RSSHub %s 失敗: %v\n", rssURL, err)
			continue
		}
		var items []model.NewsItem
		for i, entry := range feed.Items {
			if i >= 20 {
				break
			}
			title := CleanJin10Text(entry.Title)
			content := CleanJin10Text(entry.Description)
			if content == "" {
				content = title
			}
			if len(title) < 5 {
				continue
			}

			llm := enhanceFn(title, content)
			translatedTitle := llm.TranslatedTitle
			if translatedTitle == "" {
				translatedTitle = title
			}

			pubDate := ""
			if entry.Published != "" {
				pubDate = entry.Published
			}

			items = append(items, model.NewsItem{
				Title:           title,
				TranslatedTitle: translatedTitle,
				Link:            entry.Link,
				Snippet:         coalesce(llm.TranslatedSnippet, content),
				OriginalContent: content,
				PubDate:         pubDate,
				Source:          "金十數據",
				SourceColor:     "#c8a000",
				Category:        coalesce(llm.Category, "other"),
			})
		}
		if len(items) > 0 {
			return items
		}
	}
	return nil
}

func fetchJin10DirectScrape(enhanceFn func(string, string) model.LLMEnhanceResult) []model.NewsItem {
	fmt.Println("[Jin10] 改用直接爬取首頁快訊...")
	body, err := FetchURL("https://www.jin10.com/", nil)
	if err != nil {
		fmt.Printf("[Jin10] 直接爬取失敗: %v\n", err)
		return nil
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil
	}

	var items []model.NewsItem
	doc.Find(".jin-flash-item, .flash-item, .news-item, li.item").Each(func(i int, s *goquery.Selection) {
		if i >= 20 {
			return
		}
		text := CleanJin10Text(s.Text())
		if utf8.RuneCountInString(text) < 5 {
			return
		}
		titlePart := text
		if utf8.RuneCountInString(titlePart) > 80 {
			runes := []rune(titlePart)
			titlePart = string(runes[:80])
		}
		llm := enhanceFn(titlePart, text)
		translatedTitle := llm.TranslatedTitle
		if translatedTitle == "" {
			translatedTitle = titlePart
		}
		snippet := llm.TranslatedSnippet
		if snippet == "" {
			snippet = text
		}
		cat := llm.Category
		if cat == "" {
			cat = "other"
		}

		items = append(items, model.NewsItem{
			Title:           titlePart,
			TranslatedTitle: translatedTitle,
			Link:            "https://www.jin10.com/",
			Snippet:         snippet,
			OriginalContent: text,
			PubDate:         time.Now().UTC().Format(time.RFC3339),
			Source:          "金十數據",
			SourceColor:     "#c8a000",
			Category:        cat,
		})
	})
	return items
}

// deduplicateNews removes duplicate entries by title.
func deduplicateNews(items []model.NewsItem) []model.NewsItem {
	seen := make(map[string]bool)
	unique := make([]model.NewsItem, 0, len(items))
	for _, item := range items {
		key := item.Title
		if !seen[key] {
			seen[key] = true
			unique = append(unique, item)
		}
	}
	return unique
}

func coalesce(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
