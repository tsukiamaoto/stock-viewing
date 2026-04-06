package crawler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"stock-viewing-backend/internal/logger"
	"github.com/chromedp/chromedp"
)

// FetchCMoneyRealtime uses headless Chrome to bypass Nuxt rendering and fetch real posts for a given symbol.
func FetchCMoneyRealtime(symbols string) []map[string]interface{} {
	var allPosts []map[string]interface{}
	symbolList := strings.Split(symbols, ",")

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	// limit entire crawler loop
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for _, symbol := range symbolList {
		symbol = strings.TrimSpace(symbol)
		if symbol == "" {
			continue
		}

		var res string
		url := "https://www.cmoney.tw/forum/stock/" + symbol

		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.WaitVisible("article", chromedp.ByQuery),
			chromedp.Sleep(6 * time.Second), // Allow data to settle, skip skeletons
			chromedp.Evaluate(`
				Array.from(document.querySelectorAll('article, .article-card'))
					.map(el => el.innerText.trim())
					.filter(text => text.length > 10)
					.map(text => "|||" + text)
					.join('###');
			`, &res),
		)
		


		if err != nil {
			logger.Crawler().Warn("CMoney proxy failure", "symbol", symbol, "error", err)
			logger.RecordFailure("CMoney", 1)
			continue // skip on timeout or err
		}

		// parse pseudo json / separator string
		blocks := strings.Split(res, "###")
		count := 0
		for _, b := range blocks {
			if len(b) < 10 || count >= 30 {
				continue
			}
			parts := strings.SplitN(b, "|||", 2)
			title := ""
			content := strings.TrimSpace(parts[0])
			if len(parts) == 2 {
				// parts[0] is the text before ||| (usually empty since ||| is prepended)
				// parts[1] is the actual post content
				if strings.TrimSpace(parts[0]) != "" {
					title = strings.TrimSpace(parts[0])
				}
				content = strings.TrimSpace(parts[1])
			}
			
			// Fallback title to first 20 chars of content if empty
			if title == "" || title == "爆料討論" {
				runes := []rune(content)
				if len(runes) > 20 {
					title = string(runes[:20]) + "..."
				} else {
					title = content
				}
			}
			title = fmt.Sprintf("[%s] %s", symbol, title)
			if contentRunes := []rune(content); len(contentRunes) > 300 {
				content = string(contentRunes[:300]) + "..."
			}

			post := map[string]interface{}{
				"title":       title,
				"link":        url,
				"source":      "CMoney 同學會",
				"sourceColor": "#f7931e",
				"pubDate":     time.Now().Format(time.RFC3339),
				"category":    "股友",
				"snippet":     content,
				"comments":    []map[string]string{}, // Can be expanded
				"symbol":      symbol, // ADD SYMBOL HERE
			}
			allPosts = append(allPosts, post)
			count++
		}
		if count > 0 {
			logger.RecordSuccess("CMoney", count)
			logger.Crawler().Info("CMoney fetch", "symbol", symbol, "count", count)
		} else {
			logger.RecordFailure("CMoney", 1)
		}
	}

	return allPosts
}
