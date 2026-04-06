package service

import (
	"log"
	"strings"
	"sync"
)

var (
	// activeSymbols stores the symbols requested by frontend via CMoney/PTT API
	activeSymbols sync.Map
)

// RegisterWatchlistSymbols takes a comma-separated list of symbols and
// ensures they are tracked for background crawling. Returns true if ANY new symbol was added.
func RegisterWatchlistSymbols(symbols string) bool {
	if symbols == "" {
		return false
	}
	symbolList := strings.Split(symbols, ",")
	isNew := false
	for _, sym := range symbolList {
		sym = strings.TrimSpace(sym)
		if sym != "" {
			_, loaded := activeSymbols.LoadOrStore(sym, true)
			if !loaded {
				log.Printf("[Tracking] Registered new symbol for background crawl: %s\n", sym)
				isNew = true
			}
		}
	}
	return isNew
}

// GetActiveSymbols returns a slice of all tracked symbols.
func GetActiveSymbols() []string {
	var symbols []string
	activeSymbols.Range(func(key, value interface{}) bool {
		if sym, ok := key.(string); ok {
			symbols = append(symbols, sym)
		}
		return true // continue iteration
	})
	return symbols
}
