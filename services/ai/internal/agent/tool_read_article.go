package agent

import (
	"context"
	"encoding/json"
	"fmt"

	wikiPb "github.com/TenshoOHASHI/knowhub/proto/wiki"
)

type readArticleInput struct {
	ArticleID string `json:"article_id"`
}

type ReadArticleTool struct {
	wikiClient wikiPb.WikiServicesClient
}

func NewReadArticleTool(wikiClient wikiPb.WikiServicesClient) *ReadArticleTool {
	return &ReadArticleTool{wikiClient: wikiClient}
}

func (t *ReadArticleTool) Name() string { return "read_article" }

func (t *ReadArticleTool) Description() string {
	return "指定したIDの記事本文を取得します。入力: JSON {\"article_id\":\"...\"}"
}

func (t *ReadArticleTool) Run(ctx context.Context, input string) (string, error) {
	var in readArticleInput
	if err := json.Unmarshal([]byte(input), &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	resp, err := t.wikiClient.Get(ctx, &wikiPb.GetArticleRequest{Id: in.ArticleID})
	if err != nil {
		return "", fmt.Errorf("article not found: %w", err)
	}

	a := resp.Article
	result := fmt.Sprintf("# %s (ID: %s)\n\n%s", a.Title, a.Id, a.Content)
	return truncate(result, 3000), nil
}
