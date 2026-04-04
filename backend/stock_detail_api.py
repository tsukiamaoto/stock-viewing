from fastapi import APIRouter, Path
import yfinance as yf

router = APIRouter()

def _fmt_number(val, suffix=''):
    """Format large numbers to human-readable (e.g. 46.9兆)"""
    if val is None or val == 'N/A':
        return '--'
    if isinstance(val, str):
        return val
    if abs(val) >= 1e12:
        return f"{val/1e12:.2f}兆{suffix}"
    if abs(val) >= 1e8:
        return f"{val/1e8:.2f}億{suffix}"
    if abs(val) >= 1e4:
        return f"{val/1e4:.2f}萬{suffix}"
    return f"{val:,.2f}{suffix}"

def _pct(val):
    if val is None or val == 'N/A':
        return '--'
    return f"{val*100:.2f}%"

@router.get("/detail/{code}")
def get_stock_detail(code: str = Path(..., description="Stock code, e.g. 2330")):
    try:
        ticker_sym = code if "." in code else f"{code}.TW"
        ticker = yf.Ticker(ticker_sym)
        info = ticker.info

        if not info or 'shortName' not in info:
            return {"status": "error", "message": f"找不到股票代碼 {code}"}

        # Basic Info
        basic = {
            "code": code,
            "shortName": info.get("shortName", "--"),
            "longName": info.get("longName", "--"),
            "sector": info.get("sector", "--"),
            "industry": info.get("industry", "--"),
            "website": info.get("website", ""),
        }

        # Price info
        price = {
            "currentPrice": info.get("currentPrice", "--"),
            "previousClose": info.get("previousClose", "--"),
            "open": info.get("open", "--"),
            "dayHigh": info.get("dayHigh", "--"),
            "dayLow": info.get("dayLow", "--"),
            "volume": _fmt_number(info.get("volume")),
            "averageVolume": _fmt_number(info.get("averageVolume")),
            "fiftyTwoWeekHigh": info.get("fiftyTwoWeekHigh", "--"),
            "fiftyTwoWeekLow": info.get("fiftyTwoWeekLow", "--"),
            "fiftyDayAverage": info.get("fiftyDayAverage", "--"),
            "twoHundredDayAverage": info.get("twoHundredDayAverage", "--"),
            "beta": info.get("beta", "--"),
        }

        # Valuation
        valuation = {
            "marketCap": _fmt_number(info.get("marketCap")),
            "enterpriseValue": _fmt_number(info.get("enterpriseValue")),
            "trailingPE": round(info["trailingPE"], 2) if info.get("trailingPE") else "--",
            "forwardPE": round(info["forwardPE"], 2) if info.get("forwardPE") else "--",
            "priceToBook": round(info["priceToBook"], 2) if info.get("priceToBook") else "--",
            "trailingEps": info.get("trailingEps", "--"),
            "forwardEps": round(info["forwardEps"], 2) if info.get("forwardEps") else "--",
        }

        # Dividends
        dividends = {
            "dividendRate": info.get("dividendRate", "--"),
            "dividendYield": _pct(info.get("dividendYield")),
            "payoutRatio": _pct(info.get("payoutRatio")),
        }

        # Ownership
        shares_outstanding = info.get("sharesOutstanding")
        float_shares = info.get("floatShares")
        ownership = {
            "sharesOutstanding": _fmt_number(shares_outstanding),
            "floatShares": _fmt_number(float_shares),
            "heldPercentInsiders": _pct(info.get("heldPercentInsiders")),
            "heldPercentInstitutions": _pct(info.get("heldPercentInstitutions")),
        }

        # Profitability
        profitability = {
            "grossMargins": _pct(info.get("grossMargins")),
            "operatingMargins": _pct(info.get("operatingMargins")),
            "profitMargins": _pct(info.get("profitMargins")),
            "returnOnEquity": _pct(info.get("returnOnEquity")),
            "returnOnAssets": _pct(info.get("returnOnAssets")),
            "revenueGrowth": _pct(info.get("revenueGrowth")),
            "earningsGrowth": _pct(info.get("earningsQuarterlyGrowth")),
            "totalRevenue": _fmt_number(info.get("totalRevenue")),
            "netIncome": _fmt_number(info.get("netIncomeToCommon")),
        }

        # Major holders
        major_holders_data = []
        try:
            mh = ticker.major_holders
            if mh is not None and not mh.empty:
                for _, row in mh.iterrows():
                    # yfinance: column 0 = Value (e.g. "0.00022"), column 1 = Breakdown (e.g. "% of Shares Held by Insiders")
                    major_holders_data.append({
                        "value": str(row.iloc[0]) if len(row) > 0 else "",
                        "label": str(row.iloc[1]) if len(row) > 1 else "",
                    })
        except Exception:
            pass

        # Institutional holders
        inst_holders_data = []
        try:
            ih = ticker.institutional_holders
            if ih is not None and not ih.empty:
                for _, row in ih.head(15).iterrows():
                    inst_holders_data.append({
                        "holder": str(row.get("Holder", "")),
                        "shares": _fmt_number(row.get("Shares", 0)),
                        "dateReported": str(row.get("Date Reported", ""))[:10],
                        "pctHeld": _pct(row.get("pctHeld")),
                        "value": _fmt_number(row.get("Value", 0)),
                    })
        except Exception:
            pass

        return {
            "status": "success",
            "data": {
                "basic": basic,
                "price": price,
                "valuation": valuation,
                "dividends": dividends,
                "ownership": ownership,
                "profitability": profitability,
                "majorHolders": major_holders_data,
                "institutionalHolders": inst_holders_data,
            }
        }

    except Exception as e:
        print(f"[Stock Detail API] Error fetching {code}: {e}")
        return {"status": "error", "message": str(e)}
