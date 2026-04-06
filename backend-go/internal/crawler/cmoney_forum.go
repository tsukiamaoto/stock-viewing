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
			chromedp.Sleep(4*time.Second),
			// Step 1: Click all "繼續閱讀" buttons to expand full content
			chromedp.Evaluate(`
				document.querySelectorAll('.textRule__btn, [class*="readMore"]').forEach(btn => btn.click());
				'ok';
			`, &res),
			chromedp.Sleep(2*time.Second),
			// Step 2: Extract only post body content, filter ads, strip metadata
			chromedp.Evaluate(`
				(() => {
					const articles = document.querySelectorAll('article.articleContent, div.articleContent, article, .article-card');
					const results = [];
					articles.forEach(art => {
						const text = art.innerText || '';
						// Filter: skip ads / official promo posts
						if (text.includes('官方訊息') || text.includes('立即下載')) return;
						// Clone to avoid modifying the live DOM
						const clone = art.cloneNode(true);
						// Remove metadata elements
						['.articleContent__member', '.articleContent__tags', '.articleTags',
						 '.articleHavior', '.textRule__btn', '.articleContent__creator-badge',
						 '.normal__follow', '.articleContent__member-info',
						 '.articleContent__interaction', '.articleContent__report'
						].forEach(sel => {
							clone.querySelectorAll(sel).forEach(el => el.remove());
						});
						// Get the body container or fallback to cleaned clone
						const body = clone.querySelector('.articleContent__baseCont') || clone;
						let content = body.innerText.trim()
							.replace(/\n\s*\n/g, '\n')  // collapse blank lines
							.trim();
						// Strip leftover UI noise lines
						content = content.split('\n')
							.filter(line => {
								const t = line.trim();
								if (!t) return false;
								if (/^(追蹤|分享|留言|讚|打賞|繼續閱讀|Lv\.\d+|\d+分鐘前|\d+小時前|\d+天前|\d+則留言)$/.test(t)) return false;
								return true;
							})
							.join('\n');
						if (content.length > 5) {
							// Get author name for title
							const author = art.querySelector('.normal__text--sm, .articleContent__member .member__name')?.innerText?.trim() || '';
							results.push('|||' + author + '|||' + content);
						}
					});
					return results.join('###');
				})();
			`, &res),
		)
		


		if err != nil {
			logger.Crawler().Warn("CMoney proxy failure", "symbol", symbol, "error", err)
			logger.RecordFailure("CMoney", 1)
			continue // skip on timeout or err
		}

		// parse: each block is |||author|||content
		blocks := strings.Split(res, "###")
		count := 0
		for _, b := range blocks {
			if len(b) < 10 || count >= 30 {
				continue
			}
			// Format: |||author|||content
			b = strings.TrimPrefix(b, "|||")
			parts := strings.SplitN(b, "|||", 2)
			author := ""
			content := ""
			if len(parts) == 2 {
				author = strings.TrimSpace(parts[0])
				content = strings.TrimSpace(parts[1])
			} else {
				content = strings.TrimSpace(parts[0])
			}

			if len(content) < 5 {
				continue
			}

			// Build title from author or first line of content
			title := ""
			if author != "" {
				title = author
			} else {
				lines := strings.SplitN(content, "\n", 2)
				runes := []rune(lines[0])
				if len(runes) > 20 {
					title = string(runes[:20]) + "..."
				} else {
					title = lines[0]
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
				"comments":    []map[string]string{},
				"symbol":      symbol,
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
