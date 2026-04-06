package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// AppConfig holds all configuration values loaded from environment/.env file.
type AppConfig struct {
	// Supabase
	SupabaseURL string
	SupabaseKey string

	// Gemini AI
	GeminiAPIKey string
	GeminiModels []string

	// Server
	APIHost        string
	APIPort        int
	AllowedOrigins []string

	// Scheduler
	CrawlerIntervalMinutes int

	// Crawler targets
	CNNBusinessURL string
	CNNWorldURL    string

	// LLM Prompts
	SystemPrompt  string
	EnhancePrompt string
}

// Global singleton — initialized once in main.
var Cfg *AppConfig

// Load reads .env (if present) and populates Cfg.
func Load() (*AppConfig, error) {
	// Try loading .env from current dir, then parent's root .env
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	port, _ := strconv.Atoi(getEnv("API_PORT", "8000"))
	interval, _ := strconv.Atoi(getEnv("CRAWLER_INTERVAL_MINUTES", "5"))

	frontendStr := getEnv("FRONTEND_URLS", "http://localhost:5173,http://localhost:5174")
	origins := splitAndTrim(frontendStr, ",")

	modelsStr := getEnv("GEMINI_MODELS",
		"gemma-3-4b-it,gemma-3-12b-it,gemma-3-27b-it,gemma-4-26b-a4b-it,gemma-4-31b-it,gemini-3.1-flash-lite-preview,gemini-2.5-flash")
	models := splitAndTrim(modelsStr, ",")

	cfg := &AppConfig{
		SupabaseURL: getEnv("VITE_SUPABASE_URL", ""),
		SupabaseKey: getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),

		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
		GeminiModels: models,

		APIHost:        getEnv("API_HOST", "0.0.0.0"),
		APIPort:        port,
		AllowedOrigins: origins,

		CrawlerIntervalMinutes: interval,

		CNNBusinessURL: getEnv("CNN_BUSINESS_URL", "https://edition.cnn.com/business"),
		CNNWorldURL:    getEnv("CNN_WORLD_URL", "https://edition.cnn.com/world"),

		SystemPrompt: systemPrompt,
		EnhancePrompt: enhancePrompt,
	}

	Cfg = cfg
	fmt.Printf("[Config] Loaded — port=%d, models=%d, origins=%v\n", cfg.APIPort, len(cfg.GeminiModels), cfg.AllowedOrigins)
	return cfg, nil
}

// Addr returns "host:port" for the HTTP server.
func (c *AppConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.APIHost, c.APIPort)
}

// ────────────────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────────────────

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// ────────────────────────────────────────────────────────────────────
// LLM Prompts (moved from Python config.py)
// ────────────────────────────────────────────────────────────────────

const systemPrompt = `你是一個專業的金融新聞分類 AI。
使用者會提供一組新聞列表（包含標題、摘要），你需要將這些新聞分類到以下五個特定主題中：
1. "trump": 川普的相關發言或政策
2. "hormuz_iran": 荷姆茲海峽、伊朗、中東地緣政治相關新聞
3. "ai": AI 相關技術、半導體基礎設施、人工智慧新聞
4. "finance": 一般財經大盤、降息、外資報告等新聞
5. "other": 無法歸類到上述四類的新聞（前端可能不顯示）

請以 JSON 格式輸出，格式必須為：
{
  "categories": {
    "trump": [ {"title": "...", "snippet": "..."} ],
    "hormuz_iran": [ ... ],
    "ai": [ ... ],
    "finance": [ ... ]
  }
}
請只輸出合法的 JSON 字串，不要包含 ` + "```json" + ` 標籤或其他對話文字。`

const enhancePrompt = `你是一個專業的金融翻譯與分類 AI。
請閱讀以下新聞標題與內容，並完成：
1. 將新聞分類到以下五個特定主題之一："trump", "hormuz_iran", "ai", "finance", "other"
2. 將「標題」流暢地翻譯成繁體中文
3. 將「內文」精準、流暢地翻譯成繁體中文摘要。

請務必只輸出以下 JSON 格式的結果，不要帶有 markdown 或 ` + "```json" + ` 標籤：
{
  "category": "...",
  "translated_title": "...",
  "translated_snippet": "..."
}`
