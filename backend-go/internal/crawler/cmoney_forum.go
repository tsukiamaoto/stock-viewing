package crawler

import (
	"context"
	"fmt"
	"strings"
	"time"

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
			chromedp.Sleep(2 * time.Second), // Allow initial load
			chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight / 2);`, nil),
			chromedp.Sleep(1 * time.Second),
			chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil),
			chromedp.Sleep(2 * time.Second), // Allow lazy load XHR
			chromedp.Evaluate(`
				Array.from(document.querySelectorAll('article, .article-card, [class*="article"]')).map(el => {
					let title = el.querySelector('h3, .title, strong') ? el.querySelector('h3, .title, strong').innerText : '';
					let content = el.querySelector('p, .content, .text') ? el.querySelector('p, .content, .text').innerText : el.innerText;
					return title + "|||" + content;
				}).join('###');
			`, &res),
		)
		
		if err != nil {
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
			content := parts[0]
			if len(parts) == 2 && parts[0] != "" {
				title = strings.TrimSpace(parts[0])
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
			if len(content) > 300 {
				content = string([]rune(content)[:300]) + "..."
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
	}

	return allPosts
}
