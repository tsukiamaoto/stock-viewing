from supabase import create_client, Client
from config import SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY
from typing import Optional

# Singleton instance for the Supabase client
_supabase_client: Optional[Client] = None

def get_supabase() -> Client:
    """初始化並回傳 Supabase Client 單例 (Singleton)"""
    global _supabase_client
    if _supabase_client is not None:
        return _supabase_client
        
    if not SUPABASE_URL or not SUPABASE_SERVICE_ROLE_KEY or "YOUR_" in SUPABASE_SERVICE_ROLE_KEY:
        raise ValueError("Supabase URL or Key is not configured correctly in .env")
        
    _supabase_client = create_client(SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY)
    return _supabase_client

def insert_news_to_db(all_news: list[dict]):
    """
    將新聞陣列寫入 Supabase。
    """
    if not all_news:
        print("[DB] 沒有新聞可以寫入資料庫。")
        return
        
    print(f"[DB] 準備寫入 {len(all_news)} 筆新聞至 Supabase...")
    try:
        supabase = get_supabase()
        for news in all_news:
            try:
                # 使用 upsert 來避免 link 重複 (Supabase table 需設定 link 為 Unique)
                supabase.table("news").upsert({
                    "title": news.get("title", ""),
                    "translated_title": news.get("translated_title", ""),
                    "content": news.get("original_content", news.get("snippet", "")),
                    "translated_content": news.get("snippet", ""),
                    "category": news.get("category", "other"),
                    "link": news.get("link", ""),
                    "source": news.get("source", "Other"),
                    "sourceColor": news.get("sourceColor", "#666"),
                    "published_at": news.get("pubDate", "")
                }, on_conflict="link").execute()
            except Exception as e:
                print(f"[DB/Error] 寫入單筆新聞失敗: {news.get('title')} - {e}")
        print("[DB] 寫入 Supabase 完成。")
    except Exception as e:
        print(f"[DB/Error] 連線或寫入發生嚴重錯誤: {e}")
