package agent

import (
	"context"
	"fmt"

	wikiPb "github.com/TenshoOHASHI/knowhub/proto/wiki"
)

type ListArticlesTool struct {
	wikiClient wikiPb.WikiServicesClient
}

func NewListArticlesTool(wikiClient wikiPb.WikiServicesClient) *ListArticlesTool {
	return &ListArticlesTool{wikiClient: wikiClient}
}

func (t *ListArticlesTool) Name() string { return "list_articles" }

func (t *ListArticlesTool) Description() string {
	return "Wiki内の全記事のタイトル一覧を取得します。入力: \"\"（空文字）"
}

func (t *ListArticlesTool) Run(ctx context.Context, input string) (string, error) {
	articles, err := t.wikiClient.List(ctx, &wikiPb.ListArticleRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to list articles: %w", err)
	}

	if len(articles.Article) == 0 {
		return "記事はまだありません。", nil
	}

	var out string
	for i, a := range articles.Article {
		out += fmt.Sprintf("%d. [%s] ID: %s\n", i+1, a.Title, a.Id)
	}
	return truncate(out, 3000), nil
}
