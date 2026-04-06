import re

with open('tmp_cm.html', 'r', encoding='utf-8') as f:
    text = f.read()

# Let's cleanly extract Chinese text blocks that look like posts
# Nuxt puts data in arguments, but the text is still in the HTML
parts = re.split(r'\}', text)
print("Split into", len(parts))

found = 0
for p in parts:
    if len(p) > 20 and '1711' in p or '萬海' in p or '股' in p:
        # maybe extract quotes
        quotes = re.findall(r'"([^"]*?[\u4e00-\u9fa5]+[^"]*?)"', p)
        for q in quotes:
            if len(q) > 10:
                print("Text:", q)
                found += 1
            if found > 20: break
    if found > 20: break
