package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"stock-viewing-backend/internal/config"
	"stock-viewing-backend/internal/model"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// ────────────────────────────────────────────────────────────────────
// Gemini LLM Client — ports llm_classifier.py
// ────────────────────────────────────────────────────────────────────

var (
	modelIndex     uint64
	exhaustedMu    sync.Mutex
	exhaustedSet   = make(map[string]bool)
	genaiClientMu  sync.Mutex
	genaiClient    *genai.Client
	warnMissingKey sync.Once
)

func getClient() (*genai.Client, error) {
	genaiClientMu.Lock()
	defer genaiClientMu.Unlock()
	if genaiClient != nil {
		return genaiClient, nil
	}
	ctx := context.Background()
	c, err := genai.NewClient(ctx, option.WithAPIKey(config.Cfg.GeminiAPIKey))
	if err != nil {
		return nil, err
	}
	genaiClient = c
	return c, nil
}

// EnhanceNewsWithLLM classifies and translates a single news article.
// It implements model rotation with 429 exhaustion tracking, exactly like the Python version.
func EnhanceNewsWithLLM(title, snippet string) model.LLMEnhanceResult {
	fallback := model.LLMEnhanceResult{
		Category:         "other",
		TranslatedTitle:  title,
		TranslatedSnippet: snippet,
	}

	if config.Cfg.GeminiAPIKey == "" {
		warnMissingKey.Do(func() {
			fmt.Println("Warning: GEMINI_API_KEY not found in environment. Using keyword-based fallback for all news.")
		})
		return fallback
	}

	client, err := getClient()
	if err != nil {
		fmt.Printf("[LLM] Client init error: %v\n", err)
		return fallback
	}

	models := getAvailableModels()
	tried := make(map[string]bool)

	for range models {
		idx := atomic.AddUint64(&modelIndex, 1) - 1
		modelName := models[idx%uint64(len(models))]
		if tried[modelName] {
			continue
		}
		tried[modelName] = true

		userText := fmt.Sprintf("新聞標題: %s\n新聞內文: %s\n", title, snippet)
		combinedPrompt := config.Cfg.EnhancePrompt + "\n\n" + userText

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		genModel := client.GenerativeModel(modelName)
		resp, err := genModel.GenerateContent(ctx, genai.Text(combinedPrompt))
		cancel()

		if err != nil {
			if isRateLimitError(err) {
				fmt.Printf("[LLM] %s 額度用盡 (429)，暫時移除並換下一個模型...\n", modelName)
				markExhausted(modelName)
			} else {
				fmt.Printf("[LLM] %s 發生錯誤: %v，嘗試下一個模型...\n", modelName, err)
			}
			time.Sleep(500 * time.Millisecond)
			continue
		}

		text := extractResponseText(resp)
		data, err := extractJSON(text)
		if err != nil {
			fmt.Printf("[LLM] JSON parse error from %s: %v\n", modelName, err)
			continue
		}

		var result model.LLMEnhanceResult
		if c, ok := data["category"].(string); ok {
			result.Category = c
		}
		if t, ok := data["translated_title"].(string); ok {
			result.TranslatedTitle = t
		}
		if s, ok := data["translated_snippet"].(string); ok {
			result.TranslatedSnippet = s
		}
		if result.Category == "" {
			result.Category = "other"
		}
		if result.TranslatedTitle == "" {
			result.TranslatedTitle = title
		}
		if result.TranslatedSnippet == "" {
			result.TranslatedSnippet = snippet
		}
		return result
	}

	fmt.Println("[LLM] 所有可用模型均失敗，回傳原文。")
	return fallback
}

// CategorizeNewsWithLLM classifies a batch of news articles (for /categorize endpoint).
func CategorizeNewsWithLLM(newsList []model.NewsItem) map[string]interface{} {
	if config.Cfg.GeminiAPIKey == "" {
		return mockFallbackClassifier(newsList)
	}

	client, err := getClient()
	if err != nil {
		return mockFallbackClassifier(newsList)
	}

	idx := atomic.AddUint64(&modelIndex, 1) - 1
	models := config.Cfg.GeminiModels
	currentModel := models[idx%uint64(len(models))]

	var sb strings.Builder
	sb.WriteString("以下是待分類的新聞列表：\n")
	for i, n := range newsList {
		sb.WriteString(fmt.Sprintf("[%d] 標題: %s\n摘要: %s\n來源: %s\n\n", i+1, n.Title, n.Snippet, n.Source))
	}

	combinedPrompt := config.Cfg.SystemPrompt + "\n\n" + sb.String()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	m := client.GenerativeModel(currentModel)
	resp, err := m.GenerateContent(ctx, genai.Text(combinedPrompt))
	if err != nil {
		fmt.Printf("[LLM] Classification Error: %v\n", err)
		return mockFallbackClassifier(newsList)
	}

	text := extractResponseText(resp)
	data, err := extractJSON(text)
	if err != nil {
		fmt.Printf("[LLM] Classification JSON parse error: %v\n", err)
		return mockFallbackClassifier(newsList)
	}
	return data
}

// ────────────────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────────────────

func getAvailableModels() []string {
	exhaustedMu.Lock()
	defer exhaustedMu.Unlock()

	all := config.Cfg.GeminiModels
	available := make([]string, 0, len(all))
	for _, m := range all {
		if !exhaustedSet[m] {
			available = append(available, m)
		}
	}
	if len(available) == 0 {
		fmt.Println("[LLM] All models exhausted. Clearing and retrying full list.")
		exhaustedSet = make(map[string]bool)
		return all
	}
	return available
}

func markExhausted(name string) {
	exhaustedMu.Lock()
	defer exhaustedMu.Unlock()
	exhaustedSet[name] = true
}

func isRateLimitError(err error) bool {
	msg := strings.ToLower(err.Error())
	keywords := []string{"429", "resource_exhausted", "quota", "rate limit", "too many requests"}
	for _, kw := range keywords {
		if strings.Contains(msg, kw) {
			return true
		}
	}
	return false
}

// extractJSON finds the first complete JSON object from LLM output using bracket-depth tracking.
func extractJSON(text string) (map[string]interface{}, error) {
	start := strings.Index(text, "{")
	if start == -1 {
		return nil, fmt.Errorf("no JSON object found: '{' is missing")
	}

	depth := 0
	for i := start; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(text[start:i+1]), &result); err != nil {
					return nil, err
				}
				return result, nil
			}
		}
	}
	return nil, fmt.Errorf("no complete JSON object found: unmatched '{'")
}

func extractResponseText(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		sb.WriteString(fmt.Sprintf("%v", part))
	}
	return strings.TrimSpace(sb.String())
}

func mockFallbackClassifier(newsList []model.NewsItem) map[string]interface{} {
	result := map[string]interface{}{
		"categories": map[string][]map[string]string{
			"trump":       {},
			"hormuz_iran": {},
			"ai":          {},
			"finance":     {},
		},
	}
	cats := result["categories"].(map[string][]map[string]string)

	for _, n := range newsList {
		text := n.Title + " " + n.Snippet
		entry := map[string]string{"title": n.Title, "snippet": n.Snippet}

		switch {
		case containsAny(text, "川普", "Trump"):
			cats["trump"] = append(cats["trump"], entry)
		case containsAny(text, "伊朗", "荷姆茲", "中東"):
			cats["hormuz_iran"] = append(cats["hormuz_iran"], entry)
		case containsAny(text, "AI", "半導體", "晶片", "人工智慧"):
			cats["ai"] = append(cats["ai"], entry)
		default:
			cats["finance"] = append(cats["finance"], entry)
		}
	}
	return result
}

func containsAny(text string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}
