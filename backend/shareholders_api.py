"""
Scrapes shareholder distribution data (集保戶股權分散表)
from norway.twsthr.info for a given stock code.
"""
from fastapi import APIRouter, Path
import requests
from bs4 import BeautifulSoup

router = APIRouter()

HEADERS = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
}

def _parse_summary_table(soup):
    """Parse table[9] — the weekly summary (資料日期, 總股東人數, >400張 etc.)"""
    tables = soup.find_all("table")
    tbl = None
    for t in tables:
        if t.get("id") == "Details":
            rows = t.find_all("tr")
            if len(rows) > 100:
                tbl = t
                break

    if not tbl:
        return []

    rows = tbl.find_all("tr")
    results = []
    for row in rows:
        cells = row.find_all(["td", "th"])
        data = [c.get_text(strip=True) for c in cells]
        # Filter empty-prefix cells; valid data rows start with ['', '', 'YYYYMMDD', ...]
        cleaned = [d for d in data if d]
        if not cleaned or len(cleaned) < 10:
            continue
        # Skip header row
        if cleaned[0] == "資料日期":
            continue
        # Verify it looks like a date
        if len(cleaned[0]) == 8 and cleaned[0].isdigit():
            results.append({
                "date": f"{cleaned[0][:4]}/{cleaned[0][4:6]}/{cleaned[0][6:]}",
                "totalShares": cleaned[1] if len(cleaned) > 1 else "--",
                "totalHolders": cleaned[2] if len(cleaned) > 2 else "--",
                "avgShares": cleaned[3] if len(cleaned) > 3 else "--",
                "gt400Shares": cleaned[4] if len(cleaned) > 4 else "--",
                "gt400Pct": cleaned[5] if len(cleaned) > 5 else "--",
                "gt400Count": cleaned[6] if len(cleaned) > 6 else "--",
                "range400_600": cleaned[7] if len(cleaned) > 7 else "--",
                "range600_800": cleaned[8] if len(cleaned) > 8 else "--",
                "range800_1000": cleaned[9] if len(cleaned) > 9 else "--",
                "gt1000Count": cleaned[10] if len(cleaned) > 10 else "--",
                "gt1000Pct": cleaned[11] if len(cleaned) > 11 else "--",
                "closePrice": cleaned[12] if len(cleaned) > 12 else "--",
            })
    return results


def _parse_detail_table(soup):
    """Parse table[11] — the detailed breakdown by share range for latest 3 weeks."""
    tables = soup.find_all("table")
    detail_tbl = None
    for t in tables:
        if t.get("id") == "details":  # lowercase 'details'
            detail_tbl = t
            break

    if not detail_tbl:
        return {"dates": [], "rows": []}

    rows = detail_tbl.find_all("tr")
    if len(rows) < 3:
        return {"dates": [], "rows": []}

    # Row 1: dates
    date_cells = rows[1].find_all(["td", "th"])
    dates_raw = [c.get_text(strip=True) for c in date_cells]
    dates = [d for d in dates_raw if d and len(d) == 8 and d.isdigit()]
    formatted_dates = [f"{d[:4]}/{d[4:6]}/{d[6:]}" for d in dates]

    # Row 2: sub-headers (持股張數分級, 人數, 張數, 百分比% ...)
    # Rows 3+: data
    detail_rows = []
    for row in rows[3:]:
        cells = row.find_all(["td", "th"])
        data = [c.get_text(strip=True) for c in cells]
        cleaned = [d for d in data if d]
        if not cleaned:
            continue
        if cleaned[0].startswith("*") and "以上" in cleaned[0]:
            # Summary row like "* 400 張以上"
            label = cleaned[0]
        else:
            label = cleaned[0]
        
        # Each date has 3 columns: 人數, 張數, 百分比%
        periods = []
        idx = 1
        for _ in formatted_dates:
            period = {
                "holders": cleaned[idx] if idx < len(cleaned) else "--",
                "shares": cleaned[idx+1] if idx+1 < len(cleaned) else "--",
                "pct": cleaned[idx+2] if idx+2 < len(cleaned) else "--",
            }
            periods.append(period)
            idx += 3

        detail_rows.append({
            "range": label,
            "periods": periods,
        })

    return {"dates": formatted_dates, "rows": detail_rows}


@router.get("/shareholders/{code}")
def get_shareholders(code: str = Path(..., description="Stock code, e.g. 2330")):
    """Fetch shareholder distribution data (集保戶股權分散表) for a given stock."""
    try:
        url = f"https://norway.twsthr.info/StockHolders.aspx?stock={code}"
        r = requests.get(url, headers=HEADERS, timeout=15)
        r.raise_for_status()

        soup = BeautifulSoup(r.text, "html.parser")

        summary = _parse_summary_table(soup)
        detail = _parse_detail_table(soup)

        return {
            "status": "success",
            "data": {
                "code": code,
                "summary": summary[:52],  # ~1 year of weekly data
                "detail": detail,
            }
        }
    except Exception as e:
        print(f"[Shareholders API] Error fetching {code}: {e}")
        return {"status": "error", "message": str(e)}
