import urllib.request
import re

req = urllib.request.Request('https://www.cmoney.tw/forum/stock/1711', headers={'User-Agent': 'Mozilla/5.0'})
try:
    html = urllib.request.urlopen(req).read().decode('utf-8', errors='ignore')
    
    # 1. Search for internal APIs
    apis = re.findall(r'/forum/api/[a-zA-Z0-9_]+', html)
    print("Internal APIs:", set(apis))

    # 2. See if we can find any text like "1711" in standard tags
    # Write the html to tmp_cm.html so I can inspect it
    with open('tmp_cm.html', 'w', encoding='utf-8') as f:
        f.write(html)
    print("Saved HTML to tmp_cm.html")
except Exception as e:
    print('Failed:', e)
