from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager
from apscheduler.schedulers.background import BackgroundScheduler
import uvicorn
from config import ALLOWED_ORIGINS, API_HOST, API_PORT, CRAWLER_INTERVAL_MINUTES
from database import get_supabase

from news_crawler import get_all_news_for_analysis
from quotes_api import router as quotes_router
from stock_detail_api import router as stock_detail_router
from shareholders_api import router as shareholders_router
from news_api import router as news_router

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

app.include_router(news_router, prefix="/api/news", tags=["news"])


if __name__ == "__main__":
    print(f"啟動 FastAPI 後端伺服器在 {API_HOST}:{API_PORT}...")
    uvicorn.run("main:app", host=API_HOST, port=API_PORT, reload=True)
