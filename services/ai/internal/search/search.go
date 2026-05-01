package search

import (
	"context"

	"github.com/TenshoOHASHI/knowhub/services/ai/internal/embedding"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/llm"
)

// Document は検索対象の文書を表す
type Document struct {
	ID      string
	Title   string
	Content string
}

// SearchResult は検索結果の1件を表す
type SearchResult struct {
	ArticleID      string
	Title          string
	Context        string
	RelevanceScore float64
}

// SearchEngine は検索エンジンの抽象化インターフェース
type SearchEngine interface {
	// 検索対象ドキュメントをインデックスに登録する
	Index(ctx context.Context, docs []Document) error
	// キーワード検索
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
}

// SelectEngine はエンジン名から対応する SearchEngine を生成するファクトリ
func SelectEngine(engineName string, provider llm.LLMProvider, embedder embedding.EmbeddingProvider) SearchEngine {
	switch engineName {
	case "vector":
		return NewVectorEngine(embedder)
	case "hybrid":
		return NewHybridEngine(embedder, 0.5)
	case "graph":
		return NewGraphEngine(provider)
	case "bm25":
		return NewBM25Engine()
	case "tfidf":
		return NewTFIDFEngine()
	default:
		return NewBM25Engine()
	}
}
