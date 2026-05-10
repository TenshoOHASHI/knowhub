package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

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

// parseSearxHTML はSearXNGのHTMLレスポンスをパースする（JSONが使えない場合のフォールバック）
func parseSearxHTML(body string) []searxResult {
	var results []searxResult

	// シンプルな正規表現で検索結果を抽出
	// <a class="url_header" href="...">タイトル</a>
	// <p class="content">...</p>
	titleRe := regexp.MustCompile(`<a[^>]*class="url_header"[^>]*href="([^"]+)"[^>]*>([^<]+)</a>`)
	contentRe := regexp.MustCompile(`<p class="content"[^>]*>([^<]+)</p>`)

	titleMatches := titleRe.FindAllStringSubmatch(body, -1)
	contentMatches := contentRe.FindAllStringSubmatch(body, -1)

	for i, titleMatch := range titleMatches {
		if len(titleMatch) < 3 {
			continue
		}
		resultURL := titleMatch[1]
		resultTitle := strings.TrimSpace(titleMatch[2])

		// contentを取得
		resultContent := ""
		if i < len(contentMatches) && len(contentMatches[i]) > 1 {
			resultContent = strings.TrimSpace(contentMatches[i][1])
			// HTMLタグを削除
			resultContent = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(resultContent, "")
		}

		results = append(results, searxResult{
			Title:   resultTitle,
			URL:     resultURL,
			Content: resultContent,
		})
	}

	return results
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

	var results []searxResult

	// まずJSONパースを試みる
	var searxResp searxResponse
	if err := json.Unmarshal(body, &searxResp); err != nil {
		// JSONパースに失敗した場合、HTMLパースを試みる
		if len(body) > 0 && body[0] == '<' {
			slog.Info("searxng returned HTML, parsing with regex", "url", t.searxngURL)
			results = parseSearxHTML(string(body))
			if len(results) == 0 {
				return "", fmt.Errorf("failed to parse searxng HTML response (no results found)")
			}
		} else {
			return "", fmt.Errorf("failed to parse searxng response: %w", err)
		}
	} else {
		results = searxResp.Results
	}

	if len(results) == 0 {
		return "検索結果が見つかりませんでした。", nil
	}

	var out string
	limit := len(results)
	if limit > 5 {
		limit = 5
	}
	for i := 0; i < limit; i++ {
		r := results[i]
		out += fmt.Sprintf("%d. %s\n   URL: %s\n   %s\n\n", i+1, r.Title, r.URL, r.Content)
	}
	return truncate(out, 3000), nil
}
