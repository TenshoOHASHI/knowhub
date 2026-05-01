package search

import (
	"context"
	"sort"

	"github.com/TenshoOHASHI/knowhub/services/ai/internal/embedding"
)

type HybridEngine struct {
	bm25   *BM25Engine
	vector *VectorEngine
	alpha  float64 // BM25 の重み（0.0〜1.0、デフォルト 0.5）
}

func NewHybridEngine(embedder embedding.EmbeddingProvider, alpha float64) *HybridEngine {
	return &HybridEngine{
		bm25:   NewBM25Engine(),
		vector: NewVectorEngine(embedder),
		alpha:  alpha,
	}
}

func (e *HybridEngine) Index(ctx context.Context, docs []Document) error {
	// 両方のエンジンにインデックスを構築
	if err := e.bm25.Index(ctx, docs); err != nil {
		return err
	}
	if err := e.vector.Index(ctx, docs); err != nil {
		return err
	}
	return nil
}

// min-max 正規化: スコアを 0.0 〜 1.0 の範囲にスケーリング
func normalizeScores(results []SearchResult) {
	if len(results) == 0 {
		return
	}

	// 最小値・最大値を見つける
	var min, max float64
	min = results[0].RelevanceScore
	max = results[0].RelevanceScore
	for _, r := range results {
		if r.RelevanceScore < min {
			min = r.RelevanceScore
		}
		if r.RelevanceScore > max {
			max = r.RelevanceScore
		}
	}

	// (score - min) / (max - min) で 0〜1 にする
	// max == min の場合は全て同じスコア → 0 にする
	for i := range results {
		if max == min {
			results[i].RelevanceScore = 0
		} else {
			results[i].RelevanceScore = (results[i].RelevanceScore - min) / (max - min)
		}
	}
}

func (e *HybridEngine) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// ① 両エンジンで検索（limit は多めに取る）
	bm25Results, err := e.bm25.Search(ctx, query, limit*2)
	if err != nil {
		return nil, err
	}
	vectorResults, err := e.vector.Search(ctx, query, limit*2)
	if err != nil {
		return nil, err
	}

	// ② 各エンジンのスコアを正規化
	normalizeScores(bm25Results) // 認証とセキュリティ = 1.2 -> (1.2 - 0.0) / (1.2 - 0.0) = 1.0
	normalizeScores(vectorResults)

	// ③ 記事ID → 統合スコア の map を作る
	type hybridScore struct {
		bm25Score float64
		vecScore  float64
		combined  float64
		title     string
		context   string
	}

	// ポインタ型で値を上書き
	scores := make(map[string]*hybridScore)

	// BM25 の結果を map に追加
	for _, r := range bm25Results {
		scores[r.ArticleID] = &hybridScore{
			bm25Score: r.RelevanceScore, // 正規化済み
			title:     r.Title,
			context:   r.Context,
		}
	}

	// Vector の結果を map にマージ
	for _, r := range vectorResults {
		if existing, ok := scores[r.ArticleID]; ok {
			existing.vecScore = r.RelevanceScore // 正規化済み
		} else {
			scores[r.ArticleID] = &hybridScore{
				vecScore: r.RelevanceScore,
				title:    r.Title,
				context:  r.Context,
			}
		}
	}

	// ④ 統合スコアを計算: α * BM25 + (1-α) * Vector
	results := make([]SearchResult, 0, len(scores))
	for id, s := range scores {
		s.combined = e.alpha*s.bm25Score + (1-e.alpha)*s.vecScore
		if s.combined > 0 {
			results = append(results, SearchResult{
				ArticleID:      id,
				Title:          s.title,
				Context:        s.context,
				RelevanceScore: s.combined,
			})
		}
	}

	// ⑤ 統合スコア降順ソート
	sort.Slice(results, func(i, j int) bool {
		return results[i].RelevanceScore > results[j].RelevanceScore
	})

	// ⑥ limit で切り詰め
	if limit < len(results) {
		results = results[:limit]
	}

	return results, nil
}
