package crawler

import (
	"crypto/tls"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ────────────────────────────────────────────────────────────────────
// Shared HTTP client & helpers used across all crawlers
// ────────────────────────────────────────────────────────────────────

var browserHeaders = map[string]string{
	"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Accept-Language": "en-US,en;q=0.9",
	"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
}

// insecureClient skips TLS verification (development only, mirrors Python's verify=False).
var insecureClient = &http.Client{
	Timeout: 15 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

// FetchURL performs a GET request with browser-like headers.
func FetchURL(url string, extraHeaders map[string]string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range browserHeaders {
		req.Header.Set(k, v)
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := insecureClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// ────────────────────────────────────────────────────────────────────
// Text cleaning
// ────────────────────────────────────────────────────────────────────

var wsRe = regexp.MustCompile(`\s+`)

// CleanText collapses whitespace and trims.
func CleanText(s string) string {
	return strings.TrimSpace(wsRe.ReplaceAllString(s, " "))
}

// photoCreditRe filters out image-credit lines falsely parsed as titles.
var photoCreditRe = regexp.MustCompile(`(?i)Getty Images|AFP|Reuters|AP Photo|Bloomberg|Shutterstock|LightRocket|SOPA Images|via Getty|/Getty|Alamy|iStock`)

// IsValidTitle checks whether a string looks like a real headline.
func IsValidTitle(title string) bool {
	if len(title) < 25 {
		return false
	}
	if photoCreditRe.MatchString(title) {
		return false
	}
	if len(strings.Fields(title)) < 5 {
		return false
	}
	return true
}
