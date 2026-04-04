import os
from dotenv import load_dotenv

# Load .env file from the root
dotenv_path = os.path.join(os.path.dirname(os.path.dirname(__file__)), '.env')
load_dotenv(dotenv_path)

# Supabase Configurations
SUPABASE_URL = os.environ.get("VITE_SUPABASE_URL", "")
SUPABASE_SERVICE_ROLE_KEY = os.environ.get("SUPABASE_SERVICE_ROLE_KEY", "")

# Gemini API Constants
GEMINI_API_KEY = os.getenv("GEMINI_API_KEY", "")
GEMINI_MODELS_STR = os.getenv(
    "GEMINI_MODELS",
    "gemma-3-4b-it,gemma-3-12b-it,gemma-3-27b-it,gemma-4-26b-a4b-it,gemma-4-31b-it,gemini-3.1-flash-lite-preview,gemini-2.5-flash"
)
GEMINI_MODELS = [m.strip() for m in GEMINI_MODELS_STR.split(",") if m.strip()]

# API Server Configurations
API_HOST = os.getenv("API_HOST", "0.0.0.0")
API_PORT = int(os.getenv("API_PORT", "8000"))
FRONTEND_URLS_STR = os.getenv("FRONTEND_URLS", "http://localhost:5173,http://localhost:5174")
ALLOWED_ORIGINS = [u.strip() for u in FRONTEND_URLS_STR.split(",") if u.strip()]

# Scheduler Settings (default 5 minutes)
CRAWLER_INTERVAL_MINUTES = int(os.getenv("CRAWLER_INTERVAL_MINUTES", "5"))

# Crawler Targets (can be overwritten by .env)
# Using constant strings as default to not break things, but giving user capability to overwrite in .env if needed.
CNN_BUSINESS_URL = os.getenv("CNN_BUSINESS_URL", "https://edition.cnn.com/business")
CNN_WORLD_URL = os.getenv("CNN_WORLD_URL", "https://edition.cnn.com/world")

# Prompts
SYSTEM_PROMPT = """
你是一個專業的金融新聞分類 AI。
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
請只輸出合法的 JSON 字串，不要包含 ```json 標籤或其他對話文字。
"""

ENHANCE_PROMPT = """
你是一個專業的金融翻譯與分類 AI。
請閱讀以下新聞標題與內容，並完成：
1. 將新聞分類到以下五個特定主題之一："trump", "hormuz_iran", "ai", "finance", "other"
2. 將「標題」流暢地翻譯成繁體中文
3. 將「內文」精準、流暢地翻譯成繁體中文摘要。

請務必只輸出以下 JSON 格式的結果，不要帶有 markdown 或 ```json 標籤：
{
  "category": "...",
  "translated_title": "...",
  "translated_snippet": "..."
}
"""
