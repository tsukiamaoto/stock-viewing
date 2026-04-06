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

var (
	wsRe       = regexp.MustCompile(`\s+`)
	htmlTagRe  = regexp.MustCompile(`<[^>]*>`)
	htmlEntityRe = regexp.MustCompile(`&[a-zA-Z0-9#]+;`)
)

// StripHTML removes all HTML tags and common entities from text.
func StripHTML(s string) string {
	s = htmlTagRe.ReplaceAllString(s, " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	// Remove any remaining HTML entities
	s = htmlEntityRe.ReplaceAllString(s, " ")
	return s
}

// CleanText strips HTML tags, collapses whitespace and trims.
func CleanText(s string) string {
	s = StripHTML(s)
	return strings.TrimSpace(wsRe.ReplaceAllString(s, " "))
}

// ────────────────────────────────────────────────────────────────────
// Jin10-specific cleaning
// ────────────────────────────────────────────────────────────────────

var (
	// Matches "分享收藏详情复制" prefix (with optional 分享扫码 variant)
	jin10PrefixRe = regexp.MustCompile(`^(分享扫码)?分享收藏详情复制`)
	// Matches timestamps like "17:25:03" at the start (after removing prefix)
	jin10TimeRe = regexp.MustCompile(`^\d{1,2}:\d{2}(:\d{2})?`)
	// Matches "VIP" tags
	jin10VipRe = regexp.MustCompile(`VIP快讯|VIP$`)
	// Matches garbled special symbols from HTML entities
	jin10SpecialCharsRe = regexp.MustCompile(`[◆◇●○■□▲△▼▽★☆♦♠♣♥\x{FFFD}\x{FE0F}]`)
)

// CleanJin10Text removes Jin10-specific prefixes (分享收藏详情复制, timestamps, VIP tags, special chars).
func CleanJin10Text(s string) string {
	s = CleanText(s)
	s = jin10PrefixRe.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	s = jin10TimeRe.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	s = jin10VipRe.ReplaceAllString(s, "")
	s = jin10SpecialCharsRe.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	return s
}

// ────────────────────────────────────────────────────────────────────
// Title validation
// ────────────────────────────────────────────────────────────────────

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
