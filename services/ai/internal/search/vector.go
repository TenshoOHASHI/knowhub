package search

import (
	"context"
	"log/slog"
	"math"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/TenshoOHASHI/knowhub/services/ai/internal/embedding"
)

type VectorEngine struct {
	documents  []Document
	embeddings [][]float64
	embedder   embedding.EmbeddingProvider
}

// ベクトリエンジンをコンスト関数で初期化、内部でプロバイダを内包し、インターフェースで持たせる
func NewVectorEngine(embedder embedding.EmbeddingProvider) *VectorEngine {
	return &VectorEngine{
		embedder: embedder,
	}
}

// documents -> texts -> GetEmbeddings -> api(POST /api/embed) -> embed -> caches
func (e *VectorEngine) Index(ctx context.Context, docs []Document) error {
	e.documents = docs // 元の記事

	// バッファを確保
	texts := make([]string, len(docs))
	for i, doc := range docs {
		// 各文章のタイトルと内容をスペース区切りで格納する(トークン化する際に必要)
		texts[i] = doc.Title + " " + doc.Content
	}

	// 内部のGetEmbeddingメソッドを呼び出す（外部注入したAPI呼び出し担当、ollama/openAI/deepseek）
	embeddings, err := e.embedder.GetEmbeddings(ctx, texts)
	if err != nil {
		return err
	}

	e.embeddings = embeddings
	return nil
}

// tfidf.go の cosineSimilarity（map[int]float64 版）
func cosineSimilarityVec(a, b []float64) float64 {
	// クエリベクトル [0.5, -0.3, 0.8]
	// 記事Aベクトル: [0.4, -0.2, 0.7]

	var dot, normA, normB float64
	for i, v := range a {
		if i < len(b) {
			dot += v * b[i]
		}
		normA += v * v
	}
	for _, w := range b {
		normB += w * w
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// query -> embed -> queryVec & embeds -> consine
func (e *VectorEngine) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// ① クエリをベクトル化
	queryVec, err := e.embedder.GetEmbedding(ctx, query)
	if err != nil {
		return nil, err
	}

	// ② 全記事とコサイン類似度を計算
	results := make([]SearchResult, 0, len(e.documents))
	for i, doc := range e.documents {
		score := cosineSimilarityVec(queryVec, e.embeddings[i])

		// ③ 類似度が 0 より大きければ結果に追加
		if score > 0 {
			snippet := doc.Content
			if !utf8.ValidString(snippet) {
				slog.Warn("invalid UTF-8 in article content, sanitizing", "id", doc.ID, "title", doc.Title)
				snippet = strings.ToValidUTF8(snippet, "")
			}
			if !utf8.ValidString(doc.Title) {
				slog.Warn("invalid UTF-8 in article title, sanitizing", "id", doc.ID)
			}
			if len(snippet) > 200 {
				// バイト単位で切ると多バイト文字の途中で切れる可能性がある
				// rune単位で安全に切り詰める
				runes := []rune(snippet)
				if len(runes) > 200 {
					snippet = string(runes[:200]) + "..."
				}
			}
			results = append(results, SearchResult{
				ArticleID:      doc.ID,
				Title:          strings.ToValidUTF8(doc.Title, ""),
				Context:        snippet,
				RelevanceScore: score,
			})
		}
	}

	// ④ スコア降順ソート
	// slog.Info("vector search results", "total", len(results), "query", query)
	// for idx, r := range results {
	// 	slog.Info("result", "rank", idx+1, "id", r.ArticleID, "title", r.Title, "score", r.RelevanceScore)
	// }
	sort.Slice(results, func(i, j int) bool {
		return results[i].RelevanceScore > results[j].RelevanceScore
	})

	// ⑤ limit で切り詰め
	if limit < len(results) {
		results = results[:limit]
	}

	return results, nil
}
