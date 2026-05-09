package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
	slog.Info("search_wiki: tool called",
		"input", input,
		"engineName", t.engineName,
		"provider", fmt.Sprintf("%T", t.provider),
		"embedder", fmt.Sprintf("%T", t.embedder),
	)

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

		slog.Info("search_wiki: fetched articles from wiki",
			"total_articles", len(articles.Article),
		)

		docs := make([]search.Document, 0, len(articles.Article))
		for _, a := range articles.Article {
			docs = append(docs, search.Document{ID: a.Id, Title: a.Title, Content: a.Content, Visibility: a.Visibility})
		}

		slog.Info("search_wiki: prepared documents for search",
			"docs_count", len(docs),
		)

		engine := search.SelectEngine(t.engineName, t.provider, t.embedder)
		if err := engine.Index(ctx, docs); err != nil {
			return "", fmt.Errorf("failed to build index: %w", err)
		}

		results, err = engine.Search(ctx, in.Query, in.Limit)
		if err != nil {
			return "", fmt.Errorf("search failed: %w", err)
		}

		// デバッグログ
		slog.Info("search_wiki: raw results",
			"query", in.Query,
			"results_count", len(results),
			"engine", t.engineName,
		)
		for i, r := range results {
			slog.Info("search_wiki: result",
				"rank", i+1,
				"title", r.Title,
				"score", r.RelevanceScore,
			)
		}
	}

	if len(results) == 0 {
		return "該当する記事は見つかりませんでした。", nil
	}

	// 閾値フィルタ: 関連性が低い記事を除外（ai.go の ragSourceThreshold と同じ閾値）
	filtered := make([]search.SearchResult, 0, len(results))
	threshold := ragSourceThreshold(t.engineName)
	for _, r := range results {
		if r.RelevanceScore >= threshold {
			filtered = append(filtered, r)
		}
	}

	// タイトル完全一致を優先：もし完全一致するタイトルがあれば、その記事のみを返す
	var titleMatch *search.SearchResult
	queryLower := strings.ToLower(in.Query)
	for i, r := range filtered {
		titleLower := strings.ToLower(r.Title)
		if titleLower == queryLower {
			titleMatch = &filtered[i]
			break
		}
	}

	// タイトル完全一致がある場合、その記事のみを返す
	if titleMatch != nil {
		filtered = []search.SearchResult{*titleMatch}
	}

	if len(filtered) == 0 {
		return "該当する記事は見つかりませんでした。", nil
	}

	var out string
	for i, r := range filtered {
		out += fmt.Sprintf("%d. [%s] (ID: %s, スコア: %.2f)\n%s\n\n", i+1, r.Title, r.ArticleID, r.RelevanceScore, r.Context)
	}
	return truncate(out, 3000), nil
}

// ragSourceThreshold は「RAGの根拠として表示してよい最低スコア」を返す
// ai.go の ragSourceThreshold と同じ値を維持する
func ragSourceThreshold(engineName string) float64 {
	switch engineName {
	case "vector":
		return 0.70
	case "hybrid":
		return 0.50
	case "graph":
		return 1.0
	case "tfidf":
		return 0.08
	case "bm25", "":
		return 0.01 // BM25はキーワード一致なので、閾値をほぼ0にする（RAGと同じ値）
	default:
		return 0.0
	}
}
