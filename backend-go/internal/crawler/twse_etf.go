package crawler

import (
	"fmt"
	"strings"
	"time"

	"stock-viewing-backend/internal/model"

	"github.com/PuerkitoBio/goquery"
)

func FetchTWSE_ETFNews(enhanceFn func(string, string) model.LLMEnhanceResult) []model.NewsItem {
	url := "https://www.twse.com.tw/zh/ETFortune/announcementList"
	body, err := FetchURL(url, nil)
	if err != nil {
		fmt.Printf("[TWSE ETF] 抓取失敗: %v\n", err)
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		fmt.Printf("[TWSE ETF] 解析 HTML 失敗: %v\n", err)
		return nil
	}

	var items []model.NewsItem
	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		if len(items) >= 20 {
			return
		}

		tds := s.Find("td")
		if tds.Length() >= 3 {
			// td[0] = category, td[1] = date, td[2] = title and link
			dateStr := strings.TrimSpace(tds.Eq(1).Text())
			title := strings.TrimSpace(tds.Eq(2).Text())
			
			// Try to remove newlines or extra spacing inside title
			title = CleanText(title)
			if title == "" {
				return
			}

			link, exists := tds.Eq(2).Find("a").Attr("href")
			if !exists {
				return
			}
			
			fullLink := link
			if strings.HasPrefix(link, "/") {
				fullLink = "https://www.twse.com.tw" + link
			}

			pubDate := time.Now().UTC().Format(time.RFC3339)
			if parsed, err := time.Parse("2006.01.02", dateStr); err == nil {
				pubDate = parsed.UTC().Format(time.RFC3339)
			}

			llm := enhanceFn(title, "")
			cat := llm.Category
			if cat == "" {
				cat = "other"
			}

			items = append(items, model.NewsItem{
				Title:           title,
				TranslatedTitle: title, // Already in Traditional Chinese
				Link:            fullLink,
				Snippet:         title,
				OriginalContent: title,
				PubDate:         pubDate,
				Source:          "TWSE ETF 公告",
				SourceColor:     "#008c95", // TWSE e添富 theme color
				Category:        cat,
			})
		}
	})

	fmt.Printf("[TWSE ETF] 共取得 %d 則公告\n", len(items))
	return items
}
