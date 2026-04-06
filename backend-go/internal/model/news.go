package model

// NewsItem represents a single news article — the universal shape used across
// crawler → LLM → database → API response. JSON tags match the frontend contract.
type NewsItem struct {
	Title           string `json:"title"`
	TranslatedTitle string `json:"translated_title"`
	Link            string `json:"link"`
	Snippet         string `json:"snippet"`
	OriginalContent string `json:"original_content,omitempty"`
	Category        string `json:"category"`
	PubDate         string `json:"pubDate"`
	Source          string `json:"source"`
	SourceColor     string `json:"sourceColor"`
}

// CategorizedNews is the LLM classification output.
type CategorizedNews struct {
	Categories map[string][]NewsItem `json:"categories"`
}

// LLMEnhanceResult is the response from single-article LLM enhancement.
type LLMEnhanceResult struct {
	Category         string `json:"category"`
	TranslatedTitle  string `json:"translated_title"`
	TranslatedSnippet string `json:"translated_snippet"`
}

// RSSSource defines an RSS feed to crawl.
type RSSSource struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Color string `json:"color"`
}
