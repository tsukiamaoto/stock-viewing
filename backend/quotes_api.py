from fastapi import APIRouter, Query
import yfinance as yf

router = APIRouter()

@router.get("/index")
def get_index_quote(yf_symbol: str = Query(..., description="Yahoo Finance symbol, e.g. ^KS11")):
    """Fetch real index data (e.g. KOSPI ^KS11) directly from Yahoo Finance."""
    try:
        ticker = yf.Ticker(yf_symbol)
        hist = ticker.history(period="10d")

        if hist.empty:
            return {"status": "error", "message": f"No data for {yf_symbol}", "data": None}

        closes = hist['Close']
        today   = float(closes.iloc[-1])
        prev    = float(closes.iloc[-2]) if len(closes) > 1 else today
        change  = today - prev
        pct     = (change / prev * 100) if prev != 0 else 0.0

        o = float(hist['Open'].iloc[-1])
        h = float(hist['High'].iloc[-1])
        l = float(hist['Low'].iloc[-1])

        return {
            "status": "success",
            "data": {
                "symbol":        yf_symbol,
                "price":         round(today, 2),
                "change":        round(change, 2),
                "changePercent": round(pct, 2),
                "open":          round(o, 2),
                "high":          round(h, 2),
                "low":           round(l, 2),
                "prevClose":     round(prev, 2),
            }
        }
    except Exception as e:
        print(f"[Index API] Error fetching {yf_symbol}: {e}")
        return {"status": "error", "message": str(e), "data": None}


@router.get("/watchlist")
def get_watchlist_quotes(symbols: str = Query(..., description="Comma separated symbols, e.g. 2330,2317")):
    sym_list = [s.strip() for s in symbols.split(',') if s.strip()]
    if not sym_list:
        return {"status": "error", "message": "No symbols provided", "data": []}
    
    res = []
    for sym in sym_list:
        try:
            # We assume Taiwan stocks end with .TW. If it already has a suffix, trust it.
            ticker_sym = sym if "." in sym else f"{sym}.TW"
            ticker = yf.Ticker(ticker_sym)
            hist = ticker.history(period="10d")
            
            if hist.empty:
                res.append({
                    "code": sym,
                    "price": "--",
                    "change": "--", "changePercent": "--",
                    "d5_change": "--", "d5_pct": "--",
                    "d7_change": "--", "d7_pct": "--",
                    "volume": "--", "open": "--", "high": "--", "low": "--"
                })
                continue
                
            closes = hist['Close']
            today_close = float(closes.iloc[-1])
            
            # Helper to calculate change & percent
            def get_stats(prev_close):
                change = today_close - prev_close
                pct = (change / prev_close) * 100 if prev_close != 0 else 0
                return f"{change:+.2f}", f"{pct:+.2f}"
            
            d1_close = float(closes.iloc[-2]) if len(closes) > 1 else today_close
            d5_close = float(closes.iloc[-6]) if len(closes) > 5 else today_close
            d7_close = float(closes.iloc[-8]) if len(closes) > 7 else today_close
            
            c1, p1 = get_stats(d1_close)
            c5, p5 = get_stats(d5_close)
            c7, p7 = get_stats(d7_close)
            
            v = int(hist['Volume'].iloc[-1])
            o = float(hist['Open'].iloc[-1])
            h = float(hist['High'].iloc[-1])
            l = float(hist['Low'].iloc[-1])
            
            res.append({
                "code": sym,
                "price": f"{today_close:.2f}",
                "volume": f"{v:,}",
                "change": c1,
                "changePercent": p1,
                "d5_change": c5,
                "d5_pct": p5,
                "d7_change": c7,
                "d7_pct": p7,
                "open": f"{o:.2f}",
                "high": f"{h:.2f}",
                "low": f"{l:.2f}"
            })
        except Exception as e:
            print(f"[Quotes API] Error fetching {sym}: {e}")
            res.append({
                "code": sym, "price": "--", "change": "--", "changePercent": "--",
                "d5_change": "--", "d5_pct": "--", "d7_change": "--", "d7_pct": "--",
                "volume": "--", "open": "--", "high": "--", "low": "--"
            })
            
    return {"status": "success", "data": res}
