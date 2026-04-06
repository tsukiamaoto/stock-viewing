package logger

import (
	"sync"
	"time"
)

type CrawlerStats struct {
	Source       string `json:"source"`
	SuccessCount int    `json:"success_count"`
	FailureCount int    `json:"failure_count"`
	LastRun      string `json:"last_run"`
}

var (
	statsMap = make(map[string]*CrawlerStats)
	statsMu  sync.RWMutex
)

// RecordSuccess increments the success count for a specific crawler source.
func RecordSuccess(source string, count int) {
	statsMu.Lock()
	defer statsMu.Unlock()
	if _, ok := statsMap[source]; !ok {
		statsMap[source] = &CrawlerStats{Source: source}
	}
	statsMap[source].SuccessCount += count
	statsMap[source].LastRun = time.Now().UTC().Format(time.RFC3339)
}

// RecordFailure increments the failure count for a specific crawler source.
func RecordFailure(source string, count int) {
	statsMu.Lock()
	defer statsMu.Unlock()
	if _, ok := statsMap[source]; !ok {
		statsMap[source] = &CrawlerStats{Source: source}
	}
	statsMap[source].FailureCount += count
	statsMap[source].LastRun = time.Now().UTC().Format(time.RFC3339)
}

// GetStats returns the current crawler statistics.
func GetStats() []CrawlerStats {
	statsMu.RLock()
	defer statsMu.RUnlock()
	
	list := make([]CrawlerStats, 0, len(statsMap))
	for _, v := range statsMap {
		list = append(list, *v)
	}
	return list
}
