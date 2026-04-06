package crawler

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// FetchPTTStockRealtime dynamically scrapes PTT Stock for a Facebook-style feed.
// Returns map so we can freely include comments without touching the core model.
func FetchPTTStockRealtime() []map[string]interface{} {
	url := "https://www.ptt.cc/bbs/Stock/index.html"
	body, err := FetchURL(url, map[string]string{
		"Cookie": "over18=1", // Probably not needed for Stock, but safe
	})
	if err != nil {
		fmt.Printf("[PTT] Index fetch failed: %v\n", err)
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		fmt.Printf("[PTT] Parse failed: %v\n", err)
		return nil
	}

	var rawLinks []struct {
		title  string
		url    string
		date   string
		author string
	}

	// Read top 15 posts (excluding stickies if possible)
	doc.Find(".r-ent").Each(func(i int, s *goquery.Selection) {
		a := s.Find(".title a")
		title := strings.TrimSpace(a.Text())
		href, exists := a.Attr("href")
		if !exists || title == "" || strings.HasPrefix(title, "[公告]") {
			return // Skip deletes/stickies
		}
		
		date := strings.TrimSpace(s.Find(".date").Text())
		author := strings.TrimSpace(s.Find(".author").Text())
		
		rawLinks = append(rawLinks, struct {
			title, url, date, author string
		}{title: title, url: "https://www.ptt.cc" + href, date: date, author: author})
	})

	// Reverse to get newest first (PTT index is chronologically top-to-bottom)
	for i, j := 0, len(rawLinks)-1; i < j; i, j = i+1, j-1 {
		rawLinks[i], rawLinks[j] = rawLinks[j], rawLinks[i]
	}

	if len(rawLinks) > 10 {
		rawLinks = rawLinks[:10] // Limit to top 10 items for speed
	}

	var items []map[string]interface{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, p := range rawLinks {
		wg.Add(1)
		go func(title, link, author, date string) {
			defer wg.Done()
			
			// Fetch individual post for snippet and top comments
			pBody, e := FetchURL(link, map[string]string{"Cookie": "over18=1"})
			var snippet string
			var comments []map[string]string // Store comments
			if e == nil {
				pDoc, e2 := goquery.NewDocumentFromReader(strings.NewReader(string(pBody)))
				if e2 == nil {
					mainContent := pDoc.Find("#main-content").Clone()
					
					// Extract comments first before removing them!
					mainContent.Find(".push").Each(func(i int, s *goquery.Selection) {
						if len(comments) >= 5 {
							return // Only take top 5 comments
						}
						pushUserId := strings.TrimSpace(s.Find(".push-userid").Text())
						pushContent := strings.TrimSpace(s.Find(".push-content").Text())
						if pushContent != "" {
							pushContent = strings.TrimPrefix(pushContent, ": ")
							comments = append(comments, map[string]string{
								"author":  pushUserId,
								"content": pushContent,
							})
						}
					})

					// Remove meta headers and pushes
					mainContent.Children().Remove() 
					snippet = CleanText(mainContent.Text())
					if len([]rune(snippet)) > 200 {
						snippet = string([]rune(snippet)[:200]) + "..."
					}
				}
			}

			mu.Lock()
			items = append(items, map[string]interface{}{
				"title":           title,
				"translated_title": title,
				"snippet":         snippet,
				"original_content": snippet,
				"link":            link,
				"category":        author, // Repurpose category for author in the feed
				"source":          "PTT 股版",
				"sourceColor":     "#2c2c2c",
				"pubDate":         time.Now().UTC().Format(time.RFC3339),
				"comments":        comments,
			})
			mu.Unlock()
		}(p.title, p.url, p.author, p.date)
		time.Sleep(50 * time.Millisecond) // avoid rate limiting
	}

	wg.Wait()
	
	fmt.Printf("[PTT] 載入 %d 篇實時文章\n", len(items))
	return items
}
