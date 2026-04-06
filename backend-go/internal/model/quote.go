package model

// IndexQuote holds a single market index quote (e.g. ^KS11, ^GSPC).
type IndexQuote struct {
	Symbol        string  `json:"symbol"`
	Price         float64 `json:"price"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"changePercent"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	PrevClose     float64 `json:"prevClose"`
}

// WatchlistQuote holds a stock in the user's watchlist with multi-period stats.
type WatchlistQuote struct {
	Code          string `json:"code"`
	Price         string `json:"price"`
	Volume        string `json:"volume"`
	Change        string `json:"change"`
	ChangePercent string `json:"changePercent"`
	D5Change      string `json:"d5_change"`
	D5Pct         string `json:"d5_pct"`
	D7Change      string `json:"d7_change"`
	D7Pct         string `json:"d7_pct"`
	Open          string `json:"open"`
	High          string `json:"high"`
	Low           string `json:"low"`
}

// EmptyWatchlistQuote returns a placeholder when data is unavailable.
func EmptyWatchlistQuote(code string) WatchlistQuote {
	return WatchlistQuote{
		Code: code, Price: "--", Volume: "--",
		Change: "--", ChangePercent: "--",
		D5Change: "--", D5Pct: "--",
		D7Change: "--", D7Pct: "--",
		Open: "--", High: "--", Low: "--",
	}
}
