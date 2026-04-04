from news_crawler import fetch_macro_news
import sys
# Force stdout to UTF-8
sys.stdout.reconfigure(encoding='utf-8')
items = fetch_macro_news()
print(f"Total: {len(items)}")
from collections import Counter
sources = Counter(i["source"] for i in items)
print("By source:", dict(sources))
for i in items[:5]:
    print(f"  [{i['source']}] {i['title'][:80]}")
