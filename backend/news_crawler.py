import feedparser
import requests
import yfinance as yf
from bs4 import BeautifulSoup
from datetime import datetime, timezone
import re
import traceback
import urllib3
# 停用 SSL 警告（僅用於開發環境）
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
from llm_classifier import enhance_news_with_llm
from config import CNN_BUSINESS_URL, CNN_WORLD_URL

# ── 通用 HTTP headers：模擬瀏覽器，避免被封鎖 ─────────────────────────
BROWSER_HEADERS = {
    "User-Agent": (
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
        "AppleWebKit/537.36 (KHTML, like Gecko) "
        "Chrome/124.0.0.0 Safari/537.36"
    ),
    "Accept-Language": "en-US,en;q=0.9",
    "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
}

# ── CNN 爬取目標頁面 ──────────────────────────────────────────────────
CNN_SECTIONS = [
    {"url": CNN_BUSINESS_URL, "label": "Business"},
    {"url": CNN_WORLD_URL, "label": "World"},
]

# ── Reuters RSS 來源 ───────────────────────────────────────────────────
REUTERS_RSS_SOURCES = [
    {
        "name": "Reuters Business",
        "url": "https://news.google.com/rss/search?q=site:reuters.com+business&hl=en-US&gl=US&ceid=US:en",
        "color": "#ff8000",
    },
    {
        "name": "Reuters World",
        "url": "https://news.google.com/rss/search?q=site:reuters.com+world&hl=en-US&gl=US&ceid=US:en",
        "color": "#ff8000",
    },
    {
        "name": "Reuters Markets",
        "url": "https://news.google.com/rss/search?q=site:reuters.com+markets&hl=en-US&gl=US&ceid=US:en",
        "color": "#ff8000",
    },
]

# ── NHK RSS 來源 ────────────────────────────────────────────────────────
NHK_RSS_SOURCES = [
    {
        "name": "NHK World",
        "url": "https://www3.nhk.or.jp/rss/news/cat0.xml",
        "color": "#0068b7",
    },
    {
        "name": "NHK Business",
        "url": "https://www3.nhk.or.jp/rss/news/cat3.xml",
        "color": "#0068b7",
    },
]

# ── 其餘 RSS 來源（備援 / 補充）──────────────────────────────────────
RSS_SOURCES = REUTERS_RSS_SOURCES + NHK_RSS_SOURCES


# ─────────────────────────────────────────────────────────────────────
#  CNN 爬蟲核心
# ─────────────────────────────────────────────────────────────────────

def _clean_text(text: str) -> str:
    """移除多餘空白與 HTML 殘留字符"""
    text = re.sub(r'\s+', ' ', text).strip()
    return text



# 圖片版權常見關鍵字，用於過濾誤識別標題
_PHOTO_CREDIT_PATTERNS = re.compile(
    r"Getty Images|AFP|Reuters|AP Photo|Bloomberg|Shutterstock|"
    r"LightRocket|SOPA Images|via Getty|/Getty|Alamy|iStock",
    re.IGNORECASE,
)


def _is_valid_title(title: str) -> bool:
    """判斷文字是否為有效標題（排除圖片版權行、太短的文字）"""
    if len(title) < 25:
        return False
    if _PHOTO_CREDIT_PATTERNS.search(title):
        return False
    # 至少要有 5 個單詞才算標題
    if len(title.split()) < 5:
        return False
    return True


def _fetch_article_content(url: str) -> str:
    """前往文章網址，抓取內文並回傳前200個字元"""
    try:
        resp = requests.get(
            url,
            headers=BROWSER_HEADERS,
            timeout=10,
            allow_redirects=True,
            verify=False,
        )
        resp.raise_for_status()
        soup = BeautifulSoup(resp.text, "lxml")
        # 尋找所有段落
        paragraphs = soup.find_all("p")
        text = " ".join([p.get_text(strip=True) for p in paragraphs if p.get_text(strip=True)])
        clean_text = _clean_text(text)
        return clean_text[:200]
    except Exception as e:
        print(f"[CNN Content] 爬取內文失敗 {url}: {e}")
        return ""


def _extract_cnn_articles(html: str, section_label: str) -> list[dict]:
    """從 CNN 頁面 HTML 解析文章列表"""
    soup = BeautifulSoup(html, "lxml")
    articles = []
    seen_links = set()
    seen_titles = set()

    for tag in soup.select("a[href]"):
        href = tag.get("href", "")

        # 只取含日期路徑的正式文章連結
        if not re.search(r'/\d{4}/\d{2}/\d{2}/', href):
            continue

        # 組成完整 URL
        link = href if href.startswith("http") else f"https://www.cnn.com{href}"
        if link in seen_links:
            continue

        # 取標題文字
        title = _clean_text(tag.get_text(separator=" "))
        if not _is_valid_title(title) or title in seen_titles:
            continue

        seen_links.add(link)
        seen_titles.add(title)

        seen_links.add(link)
        seen_titles.add(title)

        # 抓取內文 (200字片段) 原文
        raw_article_content = _fetch_article_content(link)
        raw_snippet = raw_article_content if len(raw_article_content) > 10 else f"CNN {section_label} — {title[:150]}"

        # 透過 LLM 同時進行分類與雙語翻譯
        llm_result = enhance_news_with_llm(title, raw_snippet)

        articles.append({
            "title": title,
            "translated_title": llm_result.get("translated_title", title),
            "link": link,
            "snippet": llm_result.get("translated_snippet", raw_snippet),
            "original_content": raw_snippet,
            "category": llm_result.get("category", "other"),
            "pubDate": datetime.now(timezone.utc).strftime("%a, %d %b %Y %H:%M:%S +0000"),
            "source": f"CNN {section_label}",
            "sourceColor": "#cc0000",
        })

        if len(articles) >= 12:
            break

    return articles


def fetch_cnn_news() -> list[dict]:
    """
    爬取 CNN Business / Markets / Tech 頁面，回傳文章列表。
    每個 section 最多取 10 則，合計最多 30 則。
    """
    all_articles: list[dict] = []

    for section in CNN_SECTIONS:
        try:
            resp = requests.get(
                section["url"],
                headers=BROWSER_HEADERS,
                timeout=15,
                allow_redirects=True,
                verify=False,
            )
            resp.raise_for_status()
            articles = _extract_cnn_articles(resp.text, section["label"])
            all_articles.extend(articles[:10])
            print(f"[CNN] {section['label']} 抓取 {len(articles[:10])} 則")
        except Exception as e:
            print(f"[CNN] 爬取 {section['url']} 失敗: {e}")
            traceback.print_exc()

    # 去重（跨 section 可能重複）
    seen = set()
    unique = []
    for art in all_articles:
        if art["link"] not in seen:
            seen.add(art["link"])
            unique.append(art)

    return unique


# ─────────────────────────────────────────────────────────────────────
#  RSS 備援爬蟲
# ─────────────────────────────────────────────────────────────────────

def fetch_rss_news() -> list[dict]:
    """從其他 RSS 來源抓取補充新聞，並經過 LLM 分類與翻譯"""
    items: list[dict] = []
    for source in RSS_SOURCES:
        try:
            feed = feedparser.parse(source["url"])
            for entry in feed.entries[:8]:
                title = entry.get("title", "")
                link = entry.get("link", "")
                snippet = _clean_text(entry.get("summary", ""))
                
                # 呼叫 LLM 進行翻譯與分類
                llm_result = enhance_news_with_llm(title, snippet)
                
                items.append({
                    "title": title,
                    "link": link,
                    "snippet": llm_result.get("translated_snippet", snippet),
                    "original_content": snippet,
                    "pubDate": entry.get("published", ""),
                    "source": source["name"],
                    "sourceColor": source["color"],
                    "category": llm_result.get("category", "other"),
                    "translated_title": llm_result.get("translated_title", title),
                })
        except Exception as e:
            print(f"[RSS] {source['name']} 失敗: {e}")
    return items


def _fetch_rss_from_sources(sources: list[dict], limit_per_source: int = 10) -> list[dict]:
    """通用 RSS 抓取函數，可傳入任意來源清單"""
    items: list[dict] = []
    for source in sources:
        try:
            feed = feedparser.parse(source["url"])
            for entry in feed.entries[:limit_per_source]:
                title = entry.get("title", "")
                link = entry.get("link", "")
                snippet = _clean_text(entry.get("summary", ""))
                
                llm_result = enhance_news_with_llm(title, snippet)
                
                items.append({
                    "title": title,
                    "link": link,
                    "snippet": llm_result.get("translated_snippet", snippet),
                    "original_content": snippet,
                    "pubDate": entry.get("published", ""),
                    "source": source["name"],
                    "sourceColor": source["color"],
                    "category": llm_result.get("category", "other"),
                    "translated_title": llm_result.get("translated_title", title),
                })
        except Exception as e:
            print(f"[RSS] {source['name']} 失敗: {e}")
    return items


def fetch_reuters_news() -> list[dict]:
    """獨立抓取路透社新聞（Business + World + Markets）"""
    return _fetch_rss_from_sources(REUTERS_RSS_SOURCES, limit_per_source=10)


def fetch_nhk_news() -> list[dict]:
    """獨立抓取 NHK World 新聞"""
    return _fetch_rss_from_sources(NHK_RSS_SOURCES, limit_per_source=10)


# ─────────────────────────────────────────────────────────────────────
#  金十數據 快訊爬蟲
# ─────────────────────────────────────────────────────────────────────

def fetch_jin10_news() -> list[dict]:
    """
    爬取金十數據財經快訊（Flash 快訊）
    使用金十官方的非同步 API endpoint，回傳最新財經快訊
    """
    items: list[dict] = []
    
    # 金十數據 Flash 快訊 API（公開接口）
    jin10_urls = [
        "https://www.jin10.com/flash_newest.js",
        "https://flash-api.jin10.com/get_flash?channel=-9999&vip=1",
    ]
    
    # 先試 Flash API
    flash_api_url = "https://flash-api.jin10.com/get_flash"
    try:
        headers = {
            **BROWSER_HEADERS,
            "Referer": "https://www.jin10.com/",
            "Origin": "https://www.jin10.com",
        }
        params = {
            "channel": "-9999",
            "vip": "1",
        }
        resp = requests.get(flash_api_url, headers=headers, params=params, timeout=10, verify=False)
        resp.encoding = 'utf-8' # 強制使用 utf-8 避免亂碼 (Mojibake)
        if resp.status_code == 200:
            try:
                data = resp.json()
                flash_list = data.get("data", {}).get("data", []) if isinstance(data.get("data"), dict) else data.get("data", [])
                if not flash_list and isinstance(data, list):
                    flash_list = data
                    
                for item in flash_list[:20]:
                    content = item.get("data", {}).get("content", "") if isinstance(item.get("data"), dict) else str(item.get("content", ""))
                    content = _clean_text(content)
                    if not content or len(content) < 5:
                        continue
                    
                    pub_time = item.get("time", "") or datetime.now(timezone.utc).isoformat()
                    
                    # 呼叫 LLM 分類（快訊通常較短，以 content 當作 title）
                    llm_result = enhance_news_with_llm(content[:100], content)
                    
                    items.append({
                        "title": content[:100],
                        "translated_title": llm_result.get("translated_title", content[:100]),
                        "link": "https://www.jin10.com/",
                        "snippet": llm_result.get("translated_snippet", content),
                        "original_content": content,
                        "pubDate": pub_time,
                        "source": "金十數據",
                        "sourceColor": "#c8a000",
                        "category": llm_result.get("category", "other"),
                    })
            except Exception as parse_err:
                print(f"[Jin10] Flash API 解析失敗: {parse_err}")
    except Exception as e:
        print(f"[Jin10] Flash API 請求失敗: {e}")
    
    # 若 Flash API 失敗，使用 RSSHub 代理
    if not items:
        print("[Jin10] 改用 RSSHub 代理抓取...")
        rsshub_urls = [
            "https://rsshub.app/jin10",
            "https://rss.fatcat.app/jin10",
        ]
        for rss_url in rsshub_urls:
            try:
                feed = feedparser.parse(rss_url)
                if feed.entries:
                    for entry in feed.entries[:20]:
                        title = _clean_text(entry.get("title", ""))
                        content = _clean_text(entry.get("summary", "") or title)
                        link = entry.get("link", "https://www.jin10.com/")
                        
                        if not title or len(title) < 5:
                            continue
                        
                        llm_result = enhance_news_with_llm(title, content)
                        
                        items.append({
                            "title": title,
                            "translated_title": llm_result.get("translated_title", title),
                            "link": link,
                            "snippet": llm_result.get("translated_snippet", content),
                            "original_content": content,
                            "pubDate": entry.get("published", ""),
                            "source": "金十數據",
                            "sourceColor": "#c8a000",
                            "category": llm_result.get("category", "other"),
                        })
                    if items:
                        break
            except Exception as e:
                print(f"[Jin10] RSSHub {rss_url} 失敗: {e}")
    
    # 最後備援：直接爬取金十首頁快訊
    if not items:
        print("[Jin10] 改用直接爬取首頁快訊...")
        try:
            resp = requests.get("https://www.jin10.com/", headers=BROWSER_HEADERS, timeout=15, verify=False)
            resp.encoding = 'utf-8' # 強制使用 utf-8 避免亂碼
            if resp.status_code == 200:
                soup = BeautifulSoup(resp.text, "lxml")
                # 嘗試找到快訊列表
                flash_items = soup.select(".jin-flash-item, .flash-item, .news-item, li.item")
                for fi in flash_items[:20]:
                    text = _clean_text(fi.get_text(separator=" "))
                    if len(text) < 10:
                        continue
                    items.append({
                        "title": text[:100],
                        "translated_title": text[:100],
                        "link": "https://www.jin10.com/",
                        "snippet": text,
                        "original_content": text,
                        "pubDate": datetime.now(timezone.utc).isoformat(),
                        "source": "金十數據",
                        "sourceColor": "#c8a000",
                        "category": "other",
                    })
        except Exception as e:
            print(f"[Jin10] 直接爬取失敗: {e}")

    print(f"[Jin10] 共取得 {len(items)} 則快訊")
    return items


# ─────────────────────────────────────────────────────────────────────
#  對外主要介面
# ─────────────────────────────────────────────────────────────────────

def fetch_macro_news() -> list[dict]:
    """
    聚合宏觀財經新聞：
      1. CNN Business / Markets / Tech（直接爬取）
      2. Reuters / NHK World RSS（補充）
    """
    cnn_news = fetch_cnn_news()
    rss_news = fetch_rss_news()
    return cnn_news + rss_news



def fetch_symbol_news(symbol: str) -> list[dict]:
    """從 yfinance 抓取特定股票新聞"""
    if symbol == "Macro":
        return []
    try:
        ticker = yf.Ticker(symbol)
        news = ticker.news or []
        result = []
        for n in news[:5]:
            result.append({
                "title": n.get("title", ""),
                "link": n.get("link", ""),
                "snippet": n.get("summary", ""),
                "pubDate": "",
                "source": n.get("publisher", "Yahoo Finance"),
                "sourceColor": "#4338ca",
            })
        return result
    except Exception as e:
        print(f"[yfinance] {symbol} 失敗: {e}")
        return []


import os
from dotenv import load_dotenv

dotenv_path = os.path.join(os.path.dirname(os.path.dirname(__file__)), '.env')
load_dotenv(dotenv_path)

def _save_to_supabase(news_list: list[dict]):
    supabase_url = os.environ.get("VITE_SUPABASE_URL")
    supabase_key = os.environ.get("SUPABASE_SERVICE_ROLE_KEY")
    
    if not supabase_url or not supabase_key or "YOUR_" in supabase_key:
        print("[Supabase] 尚未設定 SUPABASE_SERVICE_ROLE_KEY，暫不寫入資料庫。")
        return
        
    try:
        from supabase import create_client, Client
        supabase: Client = create_client(supabase_url, supabase_key)
        
        # 先抓取資料庫中已存在的 link 避免重複塞入
        existing_res = supabase.table("news").select("link").execute()
        existing_links = {row["link"] for row in existing_res.data} if existing_res.data else set()
        
        insert_data = []
        for n in news_list:
            link = n.get("link", "")
            if link and link in existing_links:
                continue # 避免重複寫入相同內容
                
            insert_data.append({
                "title": n.get("title", ""),
                "content": n.get("original_content", n.get("snippet", "")),
                "translated_content": n.get("snippet", ""),
                "link": link,
                "source": n.get("source", ""),
                "sourceColor": n.get("sourceColor", ""),
                "translated_title": n.get("translated_title", ""),
                "category": n.get("category", "other"),
                "published_at": n.get("pubDate") or datetime.now(timezone.utc).isoformat()
            })
            
        if insert_data:
            supabase.table("news").insert(insert_data).execute()
            print(f"[Supabase] 成功寫入 {len(insert_data)} 筆新聞！")
    except Exception as e:
        print(f"[Supabase] 寫入資料庫失敗: {e}")

def get_all_news_for_analysis(symbol: str) -> list[dict]:
    """組合宏觀與個股新聞，供 LLM 分析使用，並自動寫入資料庫"""
    macro = fetch_macro_news()
    symbol_news = fetch_symbol_news(symbol)
    
    all_news = macro + symbol_news
    
    # 寫入 Supabase 資料庫
    _save_to_supabase(all_news)
        
    return all_news
