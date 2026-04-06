package crawler

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

// ────────────────────────────────────────────────────────────────────
// Yahoo Finance v8 API — replaces Python's yfinance library
// ────────────────────────────────────────────────────────────────────

const (
	yahooChartURL = "https://query2.finance.yahoo.com/v8/finance/chart/%s"
	yahooQuoteURL = "https://query2.finance.yahoo.com/v7/finance/quote"
)

// YahooChartResult holds the parsed chart API response.
type YahooChartResult struct {
	Timestamps []int64
	Opens      []float64
	Highs      []float64
	Lows       []float64
	Closes     []float64
	Volumes    []int64
}

// FetchYahooChart fetches OHLCV data for a symbol over a given period.
// period: "10d", "1mo", "3mo", etc.   interval: "1d", "1h"
func FetchYahooChart(symbol, period, interval string) (*YahooChartResult, error) {
	url := fmt.Sprintf(yahooChartURL, symbol)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("range", period)
	q.Set("interval", interval)
	q.Set("includePrePost", "false")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw struct {
		Chart struct {
			Result []struct {
				Timestamp  []int64 `json:"timestamp"`
				Indicators struct {
					Quote []struct {
						Open   []interface{} `json:"open"`
						High   []interface{} `json:"high"`
						Low    []interface{} `json:"low"`
						Close  []interface{} `json:"close"`
						Volume []interface{} `json:"volume"`
					} `json:"quote"`
				} `json:"indicators"`
			} `json:"result"`
			Error interface{} `json:"error"`
		} `json:"chart"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("yahoo chart decode: %w", err)
	}
	if len(raw.Chart.Result) == 0 {
		return nil, fmt.Errorf("no chart data for %s", symbol)
	}

	r := raw.Chart.Result[0]
	n := len(r.Timestamp)
	res := &YahooChartResult{
		Timestamps: r.Timestamp,
		Opens:      toFloat64Slice(r.Indicators.Quote[0].Open, n),
		Highs:      toFloat64Slice(r.Indicators.Quote[0].High, n),
		Lows:       toFloat64Slice(r.Indicators.Quote[0].Low, n),
		Closes:     toFloat64Slice(r.Indicators.Quote[0].Close, n),
		Volumes:    toInt64Slice(r.Indicators.Quote[0].Volume, n),
	}
	return res, nil
}

// toFloat64Slice converts []interface{} (JSON numbers) to []float64.
func toFloat64Slice(raw []interface{}, n int) []float64 {
	out := make([]float64, n)
	for i := 0; i < n && i < len(raw); i++ {
		if raw[i] == nil {
			continue
		}
		switch v := raw[i].(type) {
		case float64:
			out[i] = v
		case int:
			out[i] = float64(v)
		}
	}
	return out
}

// toInt64Slice converts []interface{} (JSON numbers) to []int64.
func toInt64Slice(raw []interface{}, n int) []int64 {
	out := make([]int64, n)
	for i := 0; i < n && i < len(raw); i++ {
		if raw[i] == nil {
			continue
		}
		switch v := raw[i].(type) {
		case float64:
			out[i] = int64(v)
		case int:
			out[i] = int64(v)
		}
	}
	return out
}

// YahooQuoteInfo holds quote-summary fields used for stock detail.
type YahooQuoteInfo struct {
	ShortName              string
	LongName               string
	Sector                 string
	Industry               string
	Website                string
	CurrentPrice           *float64
	PreviousClose          *float64
	Open                   *float64
	DayHigh                *float64
	DayLow                 *float64
	Volume                 *float64
	AverageVolume          *float64
	FiftyTwoWeekHigh       *float64
	FiftyTwoWeekLow        *float64
	FiftyDayAverage        *float64
	TwoHundredDayAverage   *float64
	Beta                   *float64
	MarketCap              *float64
	EnterpriseValue        *float64
	TrailingPE             *float64
	ForwardPE              *float64
	PriceToBook            *float64
	TrailingEps            *float64
	ForwardEps             *float64
	DividendRate           *float64
	DividendYield          *float64
	PayoutRatio            *float64
	SharesOutstanding      *float64
	FloatShares            *float64
	HeldPercentInsiders    *float64
	HeldPercentInstitutions *float64
	GrossMargins           *float64
	OperatingMargins       *float64
	ProfitMargins          *float64
	ReturnOnEquity         *float64
	ReturnOnAssets         *float64
	RevenueGrowth          *float64
	EarningsQuarterlyGrowth *float64
	TotalRevenue           *float64
	NetIncomeToCommon      *float64
}

// FetchYahooQuoteSummary fetches quote details for a symbol.
func FetchYahooQuoteSummary(symbol string) (*YahooQuoteInfo, error) {
	// Use v7 quote endpoint for rich data
	req, err := http.NewRequest(http.MethodGet, yahooQuoteURL, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("symbols", symbol)
	q.Set("fields", "shortName,longName,sector,industry,website,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,averageDailyVolume3Month,fiftyTwoWeekHigh,fiftyTwoWeekLow,fiftyDayAverage,twoHundredDayAverage,beta,marketCap,enterpriseValue,trailingPE,forwardPE,priceToBook,trailingEps,forwardEps,dividendRate,dividendYield,payoutRatio,sharesOutstanding,floatShares,heldPercentInsiders,heldPercentInstitutions,grossMargins,operatingMargins,profitMargins,returnOnEquity,returnOnAssets,revenueGrowth,earningsQuarterlyGrowth,totalRevenue,netIncomeToCommon")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw struct {
		QuoteResponse struct {
			Result []map[string]interface{} `json:"result"`
		} `json:"quoteResponse"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if len(raw.QuoteResponse.Result) == 0 {
		return nil, fmt.Errorf("no quote data for %s", symbol)
	}

	d := raw.QuoteResponse.Result[0]
	info := &YahooQuoteInfo{
		ShortName: getString(d, "shortName"),
		LongName:  getString(d, "longName"),
		Sector:    getString(d, "sector"),
		Industry:  getString(d, "industry"),
		Website:   getString(d, "website"),
	}

	info.CurrentPrice = getFloat(d, "regularMarketPrice")
	info.PreviousClose = getFloat(d, "regularMarketPreviousClose")
	info.Open = getFloat(d, "regularMarketOpen")
	info.DayHigh = getFloat(d, "regularMarketDayHigh")
	info.DayLow = getFloat(d, "regularMarketDayLow")
	info.Volume = getFloat(d, "regularMarketVolume")
	info.AverageVolume = getFloat(d, "averageDailyVolume3Month")
	info.FiftyTwoWeekHigh = getFloat(d, "fiftyTwoWeekHigh")
	info.FiftyTwoWeekLow = getFloat(d, "fiftyTwoWeekLow")
	info.FiftyDayAverage = getFloat(d, "fiftyDayAverage")
	info.TwoHundredDayAverage = getFloat(d, "twoHundredDayAverage")
	info.Beta = getFloat(d, "beta")
	info.MarketCap = getFloat(d, "marketCap")
	info.EnterpriseValue = getFloat(d, "enterpriseValue")
	info.TrailingPE = getFloat(d, "trailingPE")
	info.ForwardPE = getFloat(d, "forwardPE")
	info.PriceToBook = getFloat(d, "priceToBook")
	info.TrailingEps = getFloat(d, "trailingEps")
	info.ForwardEps = getFloat(d, "forwardEps")
	info.DividendRate = getFloat(d, "dividendRate")
	info.DividendYield = getFloat(d, "dividendYield")
	info.PayoutRatio = getFloat(d, "payoutRatio")
	info.SharesOutstanding = getFloat(d, "sharesOutstanding")
	info.FloatShares = getFloat(d, "floatShares")
	info.HeldPercentInsiders = getFloat(d, "heldPercentInsiders")
	info.HeldPercentInstitutions = getFloat(d, "heldPercentInstitutions")
	info.GrossMargins = getFloat(d, "grossMargins")
	info.OperatingMargins = getFloat(d, "operatingMargins")
	info.ProfitMargins = getFloat(d, "profitMargins")
	info.ReturnOnEquity = getFloat(d, "returnOnEquity")
	info.ReturnOnAssets = getFloat(d, "returnOnAssets")
	info.RevenueGrowth = getFloat(d, "revenueGrowth")
	info.EarningsQuarterlyGrowth = getFloat(d, "earningsQuarterlyGrowth")
	info.TotalRevenue = getFloat(d, "totalRevenue")
	info.NetIncomeToCommon = getFloat(d, "netIncomeToCommon")

	return info, nil
}

// ────────────────────────────────────────────────────────────────────
// Formatting helpers (port from Python's _fmt_number / _pct)
// ────────────────────────────────────────────────────────────────────

// FmtNumber formats large numbers to human-readable Chinese units.
func FmtNumber(v *float64) string {
	if v == nil {
		return "--"
	}
	val := *v
	abs := math.Abs(val)
	switch {
	case abs >= 1e12:
		return fmt.Sprintf("%.2f兆", val/1e12)
	case abs >= 1e8:
		return fmt.Sprintf("%.2f億", val/1e8)
	case abs >= 1e4:
		return fmt.Sprintf("%.2f萬", val/1e4)
	default:
		return fmt.Sprintf("%.2f", val)
	}
}

// FmtPct formats a ratio (0.xx) as a percentage string.
func FmtPct(v *float64) string {
	if v == nil {
		return "--"
	}
	return fmt.Sprintf("%.2f%%", (*v)*100)
}

// FmtRound formats a float to 2 decimal places, or "--" if nil.
func FmtRound(v *float64) interface{} {
	if v == nil {
		return "--"
	}
	return math.Round(*v*100) / 100
}

// ────────────────────────────────────────────────────────────────────
// JSON helpers
// ────────────────────────────────────────────────────────────────────

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return "--"
}

func getFloat(m map[string]interface{}, key string) *float64 {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	switch n := v.(type) {
	case float64:
		return &n
	case int:
		f := float64(n)
		return &f
	}
	return nil
}

// ────────────────────────────────────────────────────────────────────
// Ticker symbol helper
// ────────────────────────────────────────────────────────────────────

// ToYahooSymbol appends .TW if the code has no dot (Taiwan stock convention).
func ToYahooSymbol(code string) string {
	if strings.Contains(code, ".") {
		return code
	}
	return code + ".TW"
}
