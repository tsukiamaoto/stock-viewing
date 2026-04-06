package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	req, _ := http.NewRequest("GET", "https://www.cmoney.tw/forum/stock/1711", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	htmlStr := string(body)

	fmt.Println("Total HTML length:", len(htmlStr))
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		fmt.Println("Error goquery:", err)
		return
	}

	// Try extracting standard SSR text
	// Let's just find anything with "content" or "article" or lists
	texts := 0
	doc.Find("div").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if len(text) > 50 && strings.Contains(text, "1711") && texts < 5 {
			fmt.Printf("--- DIV %d ---\n%.200s\n", i, text)
			texts++
		}
	})
}
