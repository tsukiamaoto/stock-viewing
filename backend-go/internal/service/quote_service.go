package service

import (
	"fmt"
	"sync"

	"stock-viewing-backend/internal/crawler"
	"stock-viewing-backend/internal/model"
)

// ────────────────────────────────────────────────────────────────────
// Quote Service — index quote & watchlist batch
// ────────────────────────────────────────────────────────────────────

// GetIndexQuote fetches real-time data for a market index symbol (e.g. ^GSPC).
func GetIndexQuote(yfSymbol string) (*model.IndexQuote, error) {
	chart, err := crawler.FetchYahooChart(yfSymbol, "10d", "1d")
	if err != nil {
		return nil, err
	}
	closes := chart.Closes
	if len(closes) == 0 {
		return nil, fmt.Errorf("no data for %s", yfSymbol)
	}

	today := closes[len(closes)-1]
	prev := today
	if len(closes) > 1 {
		prev = closes[len(closes)-2]
	}
	change := today - prev
	pct := 0.0
	if prev != 0 {
		pct = change / prev * 100
	}

	n := len(closes) - 1

	return &model.IndexQuote{
		Symbol:        yfSymbol,
		Price:         round2(today),
		Change:        round2(change),
		ChangePercent: round2(pct),
		Open:          round2(chart.Opens[n]),
		High:          round2(chart.Highs[n]),
		Low:           round2(chart.Lows[n]),
		PrevClose:     round2(prev),
	}, nil
}

// GetWatchlistQuotes fetches quotes for multiple symbols concurrently.
func GetWatchlistQuotes(symbols []string) []model.WatchlistQuote {
	results := make([]model.WatchlistQuote, len(symbols))
	var wg sync.WaitGroup

	for i, sym := range symbols {
		wg.Add(1)
		go func(idx int, code string) {
			defer wg.Done()
			results[idx] = fetchSingleWatchlistQuote(code)
		}(i, sym)
	}
	wg.Wait()
	return results
}

func fetchSingleWatchlistQuote(code string) model.WatchlistQuote {
	tickerSym := crawler.ToYahooSymbol(code)
	chart, err := crawler.FetchYahooChart(tickerSym, "10d", "1d")
	if err != nil {
		fmt.Printf("[Quotes API] Error fetching %s: %v\n", code, err)
		return model.EmptyWatchlistQuote(code)
	}
	closes := chart.Closes
	if len(closes) == 0 {
		return model.EmptyWatchlistQuote(code)
	}

	todayClose := closes[len(closes)-1]
	n := len(closes) - 1

	getStat := func(prevIdx int) (string, string) {
		if prevIdx < 0 || prevIdx >= len(closes) {
			return "--", "--"
		}
		prev := closes[prevIdx]
		chg := todayClose - prev
		pct := 0.0
		if prev != 0 {
			pct = chg / prev * 100
		}
		return fmt.Sprintf("%+.2f", chg), fmt.Sprintf("%+.2f", pct)
	}

	c1, p1 := getStat(n - 1)
	c5, p5 := getStat(n - 5)
	c7, p7 := getStat(n - 7)

	vol := int64(0)
	if n < len(chart.Volumes) {
		vol = chart.Volumes[n]
	}

	return model.WatchlistQuote{
		Code:          code,
		Price:         fmt.Sprintf("%.2f", todayClose),
		Volume:        formatComma(vol),
		Change:        c1,
		ChangePercent: p1,
		D5Change:      c5,
		D5Pct:         p5,
		D7Change:      c7,
		D7Pct:         p7,
		Open:          fmt.Sprintf("%.2f", chart.Opens[n]),
		High:          fmt.Sprintf("%.2f", chart.Highs[n]),
		Low:           fmt.Sprintf("%.2f", chart.Lows[n]),
	}
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func formatComma(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
