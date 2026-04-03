from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import uvicorn
import os
from dotenv import load_dotenv

load_dotenv()

from news_crawler import get_all_news_for_analysis
from llm_classifier import categorize_news_with_llm

app = FastAPI(title="Finance AI Backend")

# 設定 CORS 允許 React 前端 (Vite 預設為 5173 埠) 存取
_frontend_urls = os.getenv("FRONTEND_URLS", "http://localhost:5173,http://localhost:5174")
ALLOWED_ORIGINS = [u.strip() for u in _frontend_urls.split(",")]

app.add_middleware(
    CORSMiddleware,
    allow_origins=ALLOWED_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/api/news/latest")
def get_latest_news(symbol: str = "Macro"):
    """
    抓取未經分類的生新聞，提供給前端的 RSS 新聞列表取代原本的資料。
    """
    raw_news = get_all_news_for_analysis(symbol)
    return {
        "status": "success",
        "data": raw_news
    }

@app.get("/api/news/categorize/{symbol}")
def categorize_news(symbol: str):
    """
    1. 爬蟲獲取新聞 (Crawler Layer)
    2. 透過 LLM 過濾與分類 (LLM Layer)
    3. 回傳分類好的結構化資料給前端 (Data response)
    """
    # 步驟 1: 從網站或 API 獲取最新的新聞 (包含宏觀與個股)
    raw_news = get_all_news_for_analysis(symbol)
    
    # 步驟 2: 交給 LLM 分析並分類成 川普、伊朗、AI、財經 等 bucket (受限於速度與 token，我們取前 20 則分析)
    categorized_data = categorize_news_with_llm(raw_news[:20])
    
    return {
        "symbol": symbol,
        "status": "success",
        "data": categorized_data
    }

if __name__ == "__main__":
    host = os.getenv("API_HOST", "0.0.0.0")
    port = int(os.getenv("API_PORT", "8000"))
    print(f"啟動 FastAPI 後端伺服器在 {host}:{port}...")
    uvicorn.run("main:app", host=host, port=port, reload=True)
