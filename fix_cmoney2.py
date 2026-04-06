import urllib.request
import re
import json

req = urllib.request.Request('https://www.cmoney.tw/forum/stock/1711', headers={'User-Agent': 'Mozilla/5.0'})
try:
    html = urllib.request.urlopen(req).read().decode('utf-8', errors='ignore')
    
    # We want to find the first occurrence of window.__NUXT__=(function(a,b,c,...){return {
    start = html.find('window.__NUXT__=(function(')
    if start != -1:
        # find 'return {'
        ret_st = html.find('return {', start)
        if ret_st != -1:
            end = html.find('}(', ret_st)
            json_str = html[ret_st+7:end].strip()
            
            # The JS object doesn't strictly have quoted keys, so it's not valid JSON
            # But the strings are fully visible
            # Let's extract any string mapped to 'Content' or 'content'
            contents = re.findall(r'\"?content\"?:\s*\"(.*?)\"', json_str, re.IGNORECASE)
            print("Contents found:", len(contents))
            for i, c in enumerate(contents[:3]):
                print(f"Content {i}:", c)
                
            titles = re.findall(r'\"?title\"?:\s*\"(.*?)\"', json_str, re.IGNORECASE)
            print("\nTitles found:", len(titles))
            for i, t in enumerate(titles[:3]):
                print(f"Title {i}:", t)
except Exception as e:
    print('Failed:', e)
