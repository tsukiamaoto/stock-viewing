package crawler

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"stock-viewing-backend/internal/model"

	"github.com/PuerkitoBio/goquery"
)

// ────────────────────────────────────────────────────────────────────
// CNN News Crawler — ports news_crawler.py's CNN section
// ────────────────────────────────────────────────────────────────────

type CNNSection struct {
	URL   string
	Label string
}

var datePathRe = regexp.MustCompile(`/\d{4}/\d{2}/\d{2}/`)

// FetchCNNNews crawls CNN Business/World pages and returns parsed articles.
func FetchCNNNews(sections []CNNSection, enhanceFn func(title, snippet string) model.LLMEnhanceResult) []model.NewsItem {
	var all []model.NewsItem
	for _, sec := range sections {
		articles := fetchCNNSection(sec, enhanceFn)
		if len(articles) > 10 {
			articles = articles[:10]
		}
		all = append(all, articles...)
		fmt.Printf("[CNN] %s 抓取 %d 則\n", sec.Label, len(articles))
	}

	// Deduplicate by link
	seen := make(map[string]bool)
	unique := make([]model.NewsItem, 0, len(all))
	for _, a := range all {
		if !seen[a.Link] {
			seen[a.Link] = true
			unique = append(unique, a)
		}
	}
	return unique
}

func fetchCNNSection(sec CNNSection, enhanceFn func(string, string) model.LLMEnhanceResult) []model.NewsItem {
	body, err := FetchURL(sec.URL, nil)
	if err != nil {
		fmt.Printf("[CNN] 爬取 %s 失敗: %v\n", sec.URL, err)
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil
	}

	var articles []model.NewsItem
	seenLinks := make(map[string]bool)
	seenTitles := make(map[string]bool)

	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		if len(articles) >= 12 {
			return
		}
		href, _ := s.Attr("href")
		if !datePathRe.MatchString(href) {
			return
		}
		link := href
		if !strings.HasPrefix(href, "http") {
			link = "https://www.cnn.com" + href
		}
		if seenLinks[link] {
			return
		}

		title := CleanText(s.Text())
		if !IsValidTitle(title) || seenTitles[title] {
			return
		}

		seenLinks[link] = true
		seenTitles[title] = true

		// Fetch article first 200 chars
		rawSnippet := fetchArticleContent(link)
		if len(rawSnippet) <= 10 {
			rawSnippet = fmt.Sprintf("CNN %s — %s", sec.Label, title)
			if len(rawSnippet) > 150 {
				rawSnippet = rawSnippet[:150]
			}
		}

		llm := enhanceFn(title, rawSnippet)

		translated := llm.TranslatedTitle
		if translated == "" {
			translated = title
		}
		snippet := llm.TranslatedSnippet
		if snippet == "" {
			snippet = rawSnippet
		}
		cat := llm.Category
		if cat == "" {
			cat = "other"
		}

		articles = append(articles, model.NewsItem{
			Title:           title,
			TranslatedTitle: translated,
			Link:            link,
			Snippet:         snippet,
			OriginalContent: rawSnippet,
			Category:        cat,
			PubDate:         time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 +0000"),
			Source:          fmt.Sprintf("CNN %s", sec.Label),
			SourceColor:     "#cc0000",
		})
	})

	return articles
}

func fetchArticleContent(url string) string {
	body, err := FetchURL(url, nil)
	if err != nil {
		fmt.Printf("[CNN Content] 爬取內文失敗 %s: %v\n", url, err)
		return ""
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return ""
	}
	var sb strings.Builder
	doc.Find("p").Each(func(_ int, s *goquery.Selection) {
		t := strings.TrimSpace(s.Text())
		if t != "" {
			sb.WriteString(t)
			sb.WriteString(" ")
		}
	})
	text := CleanText(sb.String())
	if len(text) > 200 {
		text = text[:200]
	}
	return text
}
