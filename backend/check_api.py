import urllib.request, json
url = "http://localhost:8000/api/news/latest?symbol=Macro"
with urllib.request.urlopen(url, timeout=30) as r:
    data = json.loads(r.read())
items = data.get("data", [])
print(f"Total articles: {len(items)}")
sources = {}
for i in items:
    s = i["source"]
    sources[s] = sources.get(s, 0) + 1
print("By source:", sources)
for i in items[:5]:
    print(f"  [{i['source']}] {i['title'][:70]}")
