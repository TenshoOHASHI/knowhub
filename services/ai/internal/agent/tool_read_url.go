package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	readability "github.com/go-shiori/go-readability"
)

type readURLInput struct {
	URL string `json:"url"`
}

type ReadURLTool struct{}

func NewReadURLTool() *ReadURLTool { return &ReadURLTool{} }

func (t *ReadURLTool) Name() string { return "read_url" }

func (t *ReadURLTool) Description() string {
	return "指定したURLのWebページ本文を抽出します。入力: JSON {\"url\":\"https://...\"}"
}

func (t *ReadURLTool) Run(ctx context.Context, input string) (string, error) {
	var in readURLInput
	if err := json.Unmarshal([]byte(input), &in); err != nil {
		in.URL = strings.TrimSpace(input)
	}
	if in.URL == "" {
		return "", fmt.Errorf("url is empty")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, in.URL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; KnowHubBot/1.0)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	parsedURL, err := url.Parse(in.URL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	article, err := readability.FromReader(bytes.NewReader(body), parsedURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse article: %w", err)
	}

	result := fmt.Sprintf("# %s\n\n%s", article.Title, article.TextContent)
	return truncate(result, 3000), nil
}
