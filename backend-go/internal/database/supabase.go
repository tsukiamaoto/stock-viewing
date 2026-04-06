package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"stock-viewing-backend/internal/config"
)

// ────────────────────────────────────────────────────────────────────
// Supabase REST Client (uses PostgREST API directly — no third-party SDK)
// ────────────────────────────────────────────────────────────────────

var (
	client *http.Client
	once   sync.Once
)

func httpClient() *http.Client {
	once.Do(func() {
		client = &http.Client{Timeout: 15 * time.Second}
	})
	return client
}

func baseURL() string {
	return strings.TrimRight(config.Cfg.SupabaseURL, "/") + "/rest/v1"
}

func authHeaders() map[string]string {
	return map[string]string{
		"apikey":        config.Cfg.SupabaseKey,
		"Authorization": "Bearer " + config.Cfg.SupabaseKey,
		"Content-Type":  "application/json",
	}
}

// ────────────────────────────────────────────────────────────────────
// Query helpers
// ────────────────────────────────────────────────────────────────────

// GetLatestNews fetches the latest N news entries ordered by published_at DESC.
func GetLatestNews(limit int) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/news?select=*&order=published_at.desc&limit=%d", baseURL(), limit)
	return doGet(url)
}

// GetNewsBySource fetches news matching a source keyword (ILIKE).
func GetNewsBySource(sourceKeyword string, limit int) ([]map[string]interface{}, error) {
	// Encode the `%` characters to `%25` so Cloudflare's worker doesn't throw a URIError.
	encodedPattern := url.QueryEscape("%" + sourceKeyword + "%")
	u := fmt.Sprintf("%s/news?select=*&source=ilike.%s&order=published_at.desc&limit=%d",
		baseURL(), encodedPattern, limit)
	return doGet(u)
}

// UpsertNews upserts a single news record (conflict on "link" column).
func UpsertNews(data map[string]interface{}) error {
	url := baseURL() + "/news"
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	for k, v := range authHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Prefer", "return=minimal")

	resp, err := httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase upsert error %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// InsertNewsBatch inserts multiple news records, skipping duplicates.
func InsertNewsBatch(items []map[string]interface{}) error {
	if len(items) == 0 {
		return nil
	}
	url := baseURL() + "/news"
	body, err := json.Marshal(items)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	for k, v := range authHeaders() {
		req.Header.Set(k, v)
	}
	// If the database does not have a UNIQUE constraint on 'link', 
	// using on_conflict=link will cause a 42P10 error. 
	// We use a standard insert here instead.
	req.Header.Set("Prefer", "return=minimal")

	resp, err := httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase batch insert error %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// ────────────────────────────────────────────────────────────────────
// Internal
// ────────────────────────────────────────────────────────────────────

func doGet(url string) ([]map[string]interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range authHeaders() {
		req.Header.Set(k, v)
	}

	resp, err := httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("supabase query error %d: %s", resp.StatusCode, string(b))
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}
