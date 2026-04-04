from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager
from apscheduler.schedulers.background import BackgroundScheduler
import uvicorn
from config import ALLOWED_ORIGINS, API_HOST, API_PORT, CRAWLER_INTERVAL_MINUTES
from database import get_supabase

from news_crawler import (
    get_all_news_for_analysis,
    fetch_cnn_news,
    fetch_reuters_news,
    fetch_nhk_news,
    fetch_jin10_news,
)
from llm_classifier import categorize_news_with_llm
from quotes_api import router as quotes_router
from stock_detail_api import router as stock_detail_router
from shareholders_api import router as shareholders_router

def scheduled_crawl_task():
    print("[Scheduler] 啟動每分鐘定時爬蟲任務...")
    try:
        # 定時爬取宏觀新聞並自動寫入 Supabase
        get_all_news_for_analysis("Macro")
    except Exception as e:
        print(f"[Scheduler] 定時爬蟲發生錯誤: {e}")

@asynccontextmanager
async def lifespan(app: FastAPI):
    # 建立一個背景排程器
    scheduler = BackgroundScheduler()
    scheduler.add_job(scheduled_crawl_task, 'interval', minutes=CRAWLER_INTERVAL_MINUTES)
    scheduler.start()
    print(f"[Scheduler] 排程器已啟動，設定為每 {CRAWLER_INTERVAL_MINUTES} 分鐘執行一次爬取。")
    yield
    scheduler.shutdown()
    print("[Scheduler] 排程器已關閉。")

app = FastAPI(title="Finance AI Backend", lifespan=lifespan)

app.add_middleware(
    CORSMiddleware,
    allow_origins=ALLOWED_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(quotes_router, prefix="/api/stocks", tags=["stocks"])
app.include_router(stock_detail_router, prefix="/api/stocks", tags=["stock-detail"])
app.include_router(shareholders_router, prefix="/api/stocks", tags=["shareholders"])

@app.get("/api/news/latest")
def get_latest_news(symbol: str = "Macro"):
    """
    從資料庫讀取最新的已儲存新聞，提供給前端介面，不再即時爬蟲。
    """
    try:
        supabase = get_supabase()
        
        # 抓取資料庫中最新的 50 筆
        res = supabase.table("news").select("*").order("published_at", desc=True).limit(50).execute()
        
        mapped_data = []
        for row in res.data:
            mapped_data.append({
                "title": row.get("title", ""),
                "translated_title": row.get("translated_title", ""),
                "snippet": row.get("translated_content") or row.get("content", ""), # 優先使用翻譯後內文
                "original_snippet": row.get("content", ""),
                "category": row.get("category", "other"),
                "link": row.get("link", ""),
                "source": row.get("source", "Other"),
                "sourceColor": row.get("sourceColor", "#666"),
                "pubDate": row.get("published_at", ""),
            })
            
        return {
            "status": "success",
            "data": mapped_data
        }
    except Exception as e:
        print(f"[News API] 讀取資料庫錯誤: {e}")
        return {"status": "error", "message": str(e)}

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


# ── Helper for DB fetching ────────────────────────────────────────────

def _get_news_by_source(source_keyword: str, limit: int = 15):
    try:
        supabase = get_supabase()
        
        # 模糊搜尋 source
        res = supabase.table("news").select("*").ilike("source", f"%{source_keyword}%")\
                .order("published_at", desc=True).limit(limit).execute()
        
        mapped_data = []
        for row in res.data:
            mapped_data.append({
                "title": row.get("title", ""),
                "translated_title": row.get("translated_title", ""),
                "snippet": row.get("translated_content") or row.get("content", ""),
                "original_content": row.get("content", ""),
                "category": row.get("category", "other"),
                "link": row.get("link", ""),
                "source": row.get("source", "Other"),
                "sourceColor": row.get("sourceColor", "#666"),
                "pubDate": row.get("published_at", ""),
            })
            
        return {"status": "success", "data": mapped_data}
    except Exception as e:
        print(f"[News API] 讀取 {source_keyword} DB 發生錯誤: {e}")
        return {"status": "error", "message": str(e), "data": []}

# ── 分區新聞 API Endpoints ────────────────────────────────────────────
@app.get("/api/news/cnn")
def get_cnn_news():
    """從資料庫讀取 CNN Business / World 新聞"""
    res = _get_news_by_source("CNN")
    res["source"] = "CNN"
    return res


@app.get("/api/news/reuters")
def get_reuters_news():
    """從資料庫讀取路透社新聞"""
    res = _get_news_by_source("Reuters")
    res["source"] = "Reuters"
    return res


@app.get("/api/news/nhk")
def get_nhk_news():
    """從資料庫讀取 NHK 新聞"""
    res = _get_news_by_source("NHK")
    res["source"] = "NHK"
    return res


@app.get("/api/news/jin10")
def get_jin10_news():
    """從資料庫讀取金十數據財經快訊"""
    res = _get_news_by_source("金十")
    res["source"] = "Jin10"
    return res


if __name__ == "__main__":
    print(f"啟動 FastAPI 後端伺服器在 {API_HOST}:{API_PORT}...")
    uvicorn.run("main:app", host=API_HOST, port=API_PORT, reload=True)
