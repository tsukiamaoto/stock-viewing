import yfinance as yf

symbols = ["2330", "2317"]
res = []
for sym in symbols:
    ticker = yf.Ticker(f"{sym}.TW")
    hist = ticker.history(period="10d")
    if hist.empty: continue
    
    closes = hist['Close']
    today = closes.iloc[-1]
    
    d1_close = closes.iloc[-2] if len(closes) > 1 else today
    d5_close = closes.iloc[-6] if len(closes) > 5 else today
    d7_close = closes.iloc[-8] if len(closes) > 7 else today

    def stats(target, prev):
        change = target - prev
        pct = (change / prev) * 100 if prev != 0 else 0
        return round(change, 2), round(pct, 2)
        
    c1, p1 = stats(today, d1_close)
    c5, p5 = stats(today, d5_close)
    c7, p7 = stats(today, d7_close)
    
    v = hist['Volume'].iloc[-1]
    res.append({
        "code": sym,
        "close": round(today, 2),
        "volume": int(v),
        "d1_change": c1, "d1_pct": p1,
        "d5_change": c5, "d5_pct": p5,
        "d7_change": c7, "d7_pct": p7
    })
print(res)
