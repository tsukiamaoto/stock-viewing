package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
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
	html := string(body)

	// In NUXT, we can try to extract article variables.
	// Often CMoney defines article object titles, e.g., title:"..." or content:"..."
	titleMatch := regexp.MustCompile(`(?i)title:\s*"([^"\\]*(?:\\.[^"\\]*)*)"`)
	titles := titleMatch.FindAllStringSubmatch(html, -1)
	
	fmt.Printf("Found %d titles\n", len(titles))
	for i, t := range titles {
		if i > 5 { break }
		fmt.Printf("Title %d: %s\n", i, t[1])
	}
	
	contentMatch := regexp.MustCompile(`(?i)content:\s*"([^"\\]*(?:\\.[^"\\]*)*)"`)
	contents := contentMatch.FindAllStringSubmatch(html, -1)
	
	fmt.Printf("\nFound %d contents\n", len(contents))
	for i, c := range contents {
		if i > 5 { break }
		fmt.Printf("Content %d: %s\n", i, c[1])
	}
    
    // Also try checking for "Title":"..." which might be the JSON representation
	jsonTitle := regexp.MustCompile(`(?i)"title":\s*"([^"\\]*(?:\\.[^"\\]*)*)"`)
	jts := jsonTitle.FindAllStringSubmatch(html, -1)
	fmt.Printf("\nFound %d JSON titles\n", len(jts))
	for i, t := range jts {
		if i > 5 { break }
		fmt.Printf("JSON Title %d: %s\n", i, t[1])
	}
}
