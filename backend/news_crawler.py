import feedparser
import yfinance as yf
from datetime import datetime

RSS_SOURCES = [
    {
        'name': 'CNN Money',
        'url': 'https://rss.cnn.com/rss/money_latest.rss',
        'color': '#cc0000',
    },
    {
        'name': 'Reuters',
        'url': 'https://www.reutersagency.com/feed/?best-topics=business-finance',
        'color': '#ff8000',
    },
    {
        'name': 'NHK World',
        'url': 'https://www3.nhk.or.jp/rss/news/cat0.xml',
        'color': '#0068b7',
    },
]

def fetch_macro_news() -> list:
    """
    從常規新聞來源爬取最新的財經與巨觀時事新聞
    """
    all_items = []
    
    for source in RSS_SOURCES:
        try:
            feed = feedparser.parse(source['url'])
            # 每個來源取前 10 則
            for entry in feed.entries[:10]:
                obj = {
                    "title": entry.get("title", ""),
                    "link": entry.get("link", ""),
                    "snippet": entry.get("summary", ""),
                    "pubDate": entry.get("published", ""),
                    "source": source['name'],
                    "sourceColor": source['color']
                }
                all_items.append(obj)
        except Exception as e:
            print(f"Error fetching RSS {source['name']}: {e}")
            
    return all_items

def fetch_symbol_news(symbol: str) -> list:
    """如果需要特定股票的新聞，可以從 yfinance 抓。"""
    if symbol == "Macro":
        return []
        
    try:
        ticker = yf.Ticker(symbol)
        news = ticker.news
        result = []
        for n in news[:5]:
            result.append({
                "title": n.get("title", ""),
                "link": n.get("link", ""),
                "snippet": n.get("summary", ""),
                "pubDate": "",
                "source": n.get("publisher", "Yahoo Finance"),
                "sourceColor": "#4338ca"
            })
        return result
    except Exception as e:
        print(f"Error fetching yfinance news: {e}")
        return []

def get_all_news_for_analysis(symbol: str) -> list:
    # 組合宏觀與特定個股的新聞
    macro = fetch_macro_news()
    symbol_news = fetch_symbol_news(symbol)
    return macro + symbol_news
