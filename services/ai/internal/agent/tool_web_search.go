package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type webSearchInput struct {
	Query string `json:"query"`
}

type searxResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

type searxResponse struct {
	Results []searxResult `json:"results"`
}

type WebSearchTool struct {
	searxngURL string
}

func NewWebSearchTool(searxngURL string) *WebSearchTool {
	return &WebSearchTool{searxngURL: searxngURL}
}

func (t *WebSearchTool) Name() string { return "web_search" }

func (t *WebSearchTool) Description() string {
	return "SearXNGでWeb検索を実行します。入力: JSON {\"query\":\"検索クエリ\"}"
}

func (t *WebSearchTool) Run(ctx context.Context, input string) (string, error) {

	var in webSearchInput
	if err := json.Unmarshal([]byte(input), &in); err != nil {
		in.Query = strings.TrimSpace(input)
	}
	if in.Query == "" {
		return "", fmt.Errorf("query is empty")
	}

	searchURL := fmt.Sprintf("%s/search?q=%s&format=json", t.searxngURL, url.QueryEscape(in.Query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("searxng request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var searxResp searxResponse
	if err := json.Unmarshal(body, &searxResp); err != nil {
		return "", fmt.Errorf("failed to parse searxng response: %w", err)
	}

	if len(searxResp.Results) == 0 {
		return "検索結果が見つかりませんでした。", nil
	}

	var out string
	limit := len(searxResp.Results)
	if limit > 5 {
		limit = 5
	}
	for i := 0; i < limit; i++ {
		r := searxResp.Results[i]
		out += fmt.Sprintf("%d. %s\n   URL: %s\n   %s\n\n", i+1, r.Title, r.URL, r.Content)
	}
	return truncate(out, 3000), nil
}
