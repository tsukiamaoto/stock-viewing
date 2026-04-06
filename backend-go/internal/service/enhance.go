package service

import (
	"fmt"
	"strings"

	"stock-viewing-backend/internal/model"
	"stock-viewing-backend/internal/translate"
)

// ────────────────────────────────────────────────────────────────────
// News Enhancement — Google Translate + keyword classification
// Replaces slow Gemini LLM for translation tasks
// ────────────────────────────────────────────────────────────────────

// EnhanceWithTranslate translates and classifies a news article using
// Google Translate (fast, free) + keyword-based classification.
// srcLang: "en", "ja", "zh-CN", "auto"
func EnhanceWithTranslate(srcLang string) func(title, snippet string) model.LLMEnhanceResult {
	return func(title, snippet string) model.LLMEnhanceResult {
		// Classify using keywords (instant, no API call)
		category := classifyByKeyword(title + " " + snippet)

		// Translate title
		translatedTitle := title
		if srcLang != "zh-TW" {
			if t, err := translate.ToTraditionalChinese(title, srcLang); err == nil && t != "" {
				translatedTitle = t
			} else if err != nil {
				fmt.Printf("[Translate] 標題翻譯失敗: %v\n", err)
			}
		}

		// Translate snippet
		translatedSnippet := snippet
		if srcLang != "zh-TW" && snippet != "" {
			if s, err := translate.ToTraditionalChinese(snippet, srcLang); err == nil && s != "" {
				translatedSnippet = s
			} else if err != nil {
				fmt.Printf("[Translate] 內文翻譯失敗: %v\n", err)
			}
		}

		return model.LLMEnhanceResult{
			Category:          category,
			TranslatedTitle:   translatedTitle,
			TranslatedSnippet: translatedSnippet,
		}
	}
}

// EnhanceJin10 handles Jin10 news: simplified→traditional Chinese + keyword classification.
func EnhanceJin10(title, snippet string) model.LLMEnhanceResult {
	category := classifyByKeyword(title + " " + snippet)

	// Convert simplified Chinese to traditional
	translatedTitle := title
	if t, err := translate.SimplifiedToTraditional(title); err == nil && t != "" {
		translatedTitle = t
	} else if err != nil {
		fmt.Printf("[Translate/Jin10] 標題簡轉繁失敗: %v (原文: %.30s...)\n", err, title)
	}

	translatedSnippet := snippet
	if snippet != "" {
		if s, err := translate.SimplifiedToTraditional(snippet); err == nil && s != "" {
			translatedSnippet = s
		} else if err != nil {
			fmt.Printf("[Translate/Jin10] 內文簡轉繁失敗: %v\n", err)
		}
	}

	return model.LLMEnhanceResult{
		Category:          category,
		TranslatedTitle:   translatedTitle,
		TranslatedSnippet: translatedSnippet,
	}
}

// classifyByKeyword does instant keyword-based news classification.
func classifyByKeyword(text string) string {
	lower := strings.ToLower(text)
	switch {
	case containsAny(text, "川普", "Trump", "trump", "特朗普", "關稅", "关税", "tariff", "Tariff"):
		return "trump"
	case containsAny(text, "伊朗", "Iran", "iran", "荷姆茲", "霍尔木兹", "中東", "中东", "Hormuz", "hormuz"):
		return "hormuz_iran"
	case containsAny(lower, "ai", "半導體", "半导体", "晶片", "芯片", "人工智慧", "人工智能",
		"semiconductor", "nvidia", "openai", "chatgpt", "gpu"):
		return "ai"
	case containsAny(text, "降息", "升息", "Fed", "聯準會", "联储", "GDP", "CPI",
		"央行", "利率", "匯率", "汇率", "外資", "外资", "stock", "bond",
		"S&P", "道瓊", "道琼", "納斯達克", "纳斯达克", "Wall Street"):
		return "finance"
	default:
		return "finance"
	}
}

func containsAny(text string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}
