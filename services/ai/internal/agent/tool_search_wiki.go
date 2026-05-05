package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	wikiPb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/embedding"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/llm"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/search"
)

type searchWikiInput struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

// GraphProvider はキャッシュ済み GraphEngine を取得する関数の型
type GraphProvider func(ctx context.Context) (*search.GraphEngine, error)

type SearchWikiTool struct {
	wikiClient    wikiPb.WikiServicesClient
	provider      llm.LLMProvider
	embedder      embedding.EmbeddingProvider
	engineName    string
	graphProvider GraphProvider // Graph RAG 用キャッシュプロバイダー（nil の場合は毎回構築）
}

func NewSearchWikiTool(wikiClient wikiPb.WikiServicesClient, provider llm.LLMProvider, embedder embedding.EmbeddingProvider, engineName string, graphProvider GraphProvider) *SearchWikiTool {
	return &SearchWikiTool{
		wikiClient:    wikiClient,
		provider:      provider,
		embedder:      embedder,
		engineName:    engineName,
		graphProvider: graphProvider,
	}
}

func (t *SearchWikiTool) Name() string { return "search_wiki" }

func (t *SearchWikiTool) Description() string {
	return "Wiki 内の記事を検索します。入力: JSON {\"query\":\"検索クエリ\",\"limit\":5}"
}

func (t *SearchWikiTool) Run(ctx context.Context, input string) (string, error) {
	var in searchWikiInput
	if err := json.Unmarshal([]byte(input), &in); err != nil {
		// LLMがプレーン文字列を出力した場合のフォールバック
		in.Query = strings.TrimSpace(input)
	}
	if in.Query == "" {
		return "", fmt.Errorf("query is empty")
	}
	if in.Limit <= 0 {
		in.Limit = 5
	}

	// Graph RAG の場合はキャッシュ済みグラフを使う（毎回 Index すると全記事 × LLM 呼び出しでタイムアウトする）
	var results []search.SearchResult
	if t.engineName == "graph" && t.graphProvider != nil {
		graphEngine, err := t.graphProvider(ctx) // h.ensureGraphを実行、キャッシュ済みならそれを変えす、なければ構築
		if err != nil {
			return "", fmt.Errorf("failed to get cached graph: %w", err)
		}
		results, err = graphEngine.Search(ctx, in.Query, in.Limit)
		if err != nil {
			return "", fmt.Errorf("graph search failed: %w", err)
		}
	} else {
		// Wiki から全記事取得
		articles, err := t.wikiClient.List(ctx, &wikiPb.ListArticleRequest{})
		if err != nil {
			return "", fmt.Errorf("failed to list articles: %w", err)
		}

		docs := make([]search.Document, 0, len(articles.Article))
		for _, a := range articles.Article {
			docs = append(docs, search.Document{ID: a.Id, Title: a.Title, Content: a.Content})
		}

		engine := search.SelectEngine(t.engineName, t.provider, t.embedder)
		if err := engine.Index(ctx, docs); err != nil {
			return "", fmt.Errorf("failed to build index: %w", err)
		}

		results, err = engine.Search(ctx, in.Query, in.Limit)
		if err != nil {
			return "", fmt.Errorf("search failed: %w", err)
		}
	}

	if len(results) == 0 {
		return "該当する記事は見つかりませんでした。", nil
	}

	var out string
	for i, r := range results {
		out += fmt.Sprintf("%d. [%s] (ID: %s, スコア: %.2f)\n%s\n\n", i+1, r.Title, r.ArticleID, r.RelevanceScore, r.Context)
	}
	return truncate(out, 3000), nil
}
