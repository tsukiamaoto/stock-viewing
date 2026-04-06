package translate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ────────────────────────────────────────────────────────────────────
// Google Translate (free, no API key) — fast replacement for Gemini LLM translation
// ────────────────────────────────────────────────────────────────────

var client = &http.Client{Timeout: 10 * time.Second}

var (
	rateMu      sync.Mutex
	lastRequest time.Time
)

// ToTraditionalChinese translates text to Traditional Chinese (zh-TW).
// srcLang can be "en", "ja", "zh-CN", or "auto" for auto-detection.
func ToTraditionalChinese(text, srcLang string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return text, nil
	}

	// Rate limit: max ~10 requests/second to avoid Google blocking
	rateMu.Lock()
	if time.Since(lastRequest) < 100*time.Millisecond {
		time.Sleep(100*time.Millisecond - time.Since(lastRequest))
	}
	lastRequest = time.Now()
	rateMu.Unlock()

	// Google Translate free endpoint
	apiURL := fmt.Sprintf(
		"https://translate.googleapis.com/translate_a/single?client=gtx&sl=%s&tl=zh-TW&dt=t&q=%s",
		url.QueryEscape(srcLang),
		url.QueryEscape(text),
	)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return text, fmt.Errorf("translate request creation: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return text, fmt.Errorf("translate request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return text, fmt.Errorf("translate API returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return text, fmt.Errorf("translate read body: %w", err)
	}

	return parseGoogleTranslateResponse(body)
}

// SimplifiedToTraditional converts simplified Chinese to traditional Chinese.
func SimplifiedToTraditional(text string) (string, error) {
	return ToTraditionalChinese(text, "zh-CN")
}

// EnglishToTraditionalChinese translates English to Traditional Chinese.
func EnglishToTraditionalChinese(text string) (string, error) {
	return ToTraditionalChinese(text, "en")
}

// JapaneseToTraditionalChinese translates Japanese to Traditional Chinese.
func JapaneseToTraditionalChinese(text string) (string, error) {
	return ToTraditionalChinese(text, "ja")
}

// AutoToTraditionalChinese translates any language to Traditional Chinese.
func AutoToTraditionalChinese(text string) (string, error) {
	return ToTraditionalChinese(text, "auto")
}

// parseGoogleTranslateResponse extracts the translated text from Google's response.
// Response format: [[["translated text","source text",null,null,10]],null,"en"]
func parseGoogleTranslateResponse(body []byte) (string, error) {
	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("translate JSON parse: %w", err)
	}

	// The response is a nested array: result[0] is an array of translation segments
	arr, ok := result.([]interface{})
	if !ok || len(arr) == 0 {
		return "", fmt.Errorf("unexpected translate response format")
	}

	segments, ok := arr[0].([]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected translate segments format")
	}

	var sb strings.Builder
	for _, seg := range segments {
		segArr, ok := seg.([]interface{})
		if !ok || len(segArr) == 0 {
			continue
		}
		if translated, ok := segArr[0].(string); ok {
			sb.WriteString(translated)
		}
	}

	result_text := sb.String()
	if result_text == "" {
		return "", fmt.Errorf("empty translation result")
	}
	return result_text, nil
}

// BatchTranslate translates multiple texts concurrently.
// Returns translated texts in the same order. On failure, returns original text.
func BatchTranslate(texts []string, srcLang string) []string {
	results := make([]string, len(texts))
	type result struct {
		idx  int
		text string
	}
	ch := make(chan result, len(texts))

	for i, text := range texts {
		go func(idx int, t string) {
			translated, err := ToTraditionalChinese(t, srcLang)
			if err != nil {
				fmt.Printf("[Translate] 翻譯失敗: %v, 使用原文\n", err)
				ch <- result{idx, t}
				return
			}
			ch <- result{idx, translated}
		}(i, text)
	}

	for range texts {
		r := <-ch
		results[r.idx] = r.text
	}
	return results
}
