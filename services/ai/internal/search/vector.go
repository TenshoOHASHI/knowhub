package search

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/TenshoOHASHI/knowhub/services/ai/internal/embedding"
)

type VectorEngine struct {
	documents      map[string]Document  // ID → Document
	embeddings     map[string][]float64 // 記事ID → 埋め込みベクトル
	embedder       embedding.EmbeddingProvider
	articleUpdated map[string]time.Time // 記事ID → 最終更新時刻（差分検知用）
	persistPath    string               // 永続化ファイルパス
}

// ベクトリエンジンをコンスト関数で初期化、内部でプロバイダを内包し、インターフェースで持たせる
func NewVectorEngine(embedder embedding.EmbeddingProvider) *VectorEngine {
	return &VectorEngine{
		embedder:       embedder,
		documents:      make(map[string]Document),
		embeddings:     make(map[string][]float64),
		articleUpdated: make(map[string]time.Time),
	}
}

// ============================================================
// 永続化: 埋め込みベクトルをファイルに保存/読み込み
// ============================================================

// vectorPersist は永続化用のデータ構造
type vectorPersist struct {
	Embeddings     map[string][]float64 `json:"embeddings"`
	Docs           []Document           `json:"docs"`
	ArticleUpdated map[string]string    `json:"article_updated"` // ISO8601形式
}

// SaveEmbeddings は埋め込みベクトルをJSONファイルに保存する
func (e *VectorEngine) SaveEmbeddings(path string) error {
	// 記事をスライスに変換
	docs := make([]Document, 0, len(e.documents))
	for _, doc := range e.documents {
		docs = append(docs, doc)
	}

	// articleUpdatedをISO8601形式に変換
	articleUpdated := make(map[string]string)
	for id, t := range e.articleUpdated {
		articleUpdated[id] = t.Format(time.RFC3339)
	}

	data := vectorPersist{
		Embeddings:     e.embeddings,
		Docs:           docs,
		ArticleUpdated: articleUpdated,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal vector embeddings: %w", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("write vector embeddings file: %w", err)
	}

	slog.Info("vector: saved embeddings to file", "path", path,
		"docs", len(docs),
		"embeddings", len(e.embeddings))

	return nil
}

// LoadEmbeddings はJSONファイルから埋め込みベクトルを読み込む
func (e *VectorEngine) LoadEmbeddings(path string) error {
	jsonData, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("vector: no saved embeddings found", "path", path)
			return nil // ファイルがないのはエラーではない
		}
		return fmt.Errorf("read vector embeddings file: %w", err)
	}

	var data vectorPersist
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("unmarshal vector embeddings: %w", err)
	}

	// ドキュメントを復元
	e.documents = make(map[string]Document)
	for _, doc := range data.Docs {
		e.documents[doc.ID] = doc
	}

	// 埋め込みベクトルを復元
	e.embeddings = data.Embeddings
	if e.embeddings == nil {
		e.embeddings = make(map[string][]float64)
	}

	// articleUpdatedを復元
	e.articleUpdated = make(map[string]time.Time)
	for id, ts := range data.ArticleUpdated {
		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			slog.Warn("vector: failed to parse article timestamp", "id", id, "timestamp", ts)
			continue
		}
		e.articleUpdated[id] = t
	}

	slog.Info("vector: loaded embeddings from file", "path", path,
		"docs", len(e.documents),
		"embeddings", len(e.embeddings))

	return nil
}

// ============================================================
// 差分更新対応の Index メソッド
// ============================================================

// documents -> texts -> GetEmbeddings -> api(POST /api/embed) -> embed -> caches
func (e *VectorEngine) Index(ctx context.Context, docs []Document) error {
	// 現在の記事IDセットを構築（公開記事のみ）
	currentDocIDs := make(map[string]bool)
	for _, doc := range docs {
		if doc.Visibility == "public" {
			currentDocIDs[doc.ID] = true
		}
	}

	// 削除された記事を特定 → map から削除
	for id := range e.documents {
		if !currentDocIDs[id] {
			delete(e.documents, id)
			delete(e.embeddings, id)
			delete(e.articleUpdated, id)
			slog.Info("vector: removed deleted article", "id", id)
		}
	}

	// 新規・更新された記事のみ特定
	var toProcess []Document
	for _, doc := range docs {
		if doc.Visibility != "public" {
			continue
		}

		lastUpdated, exists := e.articleUpdated[doc.ID]
		if !exists {
			// 新規記事
			toProcess = append(toProcess, doc)
		} else if !doc.UpdatedAt.IsZero() && doc.UpdatedAt.After(lastUpdated) {
			// 更新された記事
			toProcess = append(toProcess, doc)
		}
		// 既存かつ未更新の記事はスキップ（キャッシュ済み埋め込みベクトルを使用）
	}

	slog.Info("vector: incremental update",
		"total_articles", len(docs),
		"new_or_updated", len(toProcess),
		"cached", len(currentDocIDs)-len(toProcess))

	// 差分がなければ早期リターン
	if len(toProcess) == 0 {
		// ドキュメント情報だけは最新に更新（visibility変更等に対応）
		for _, doc := range docs {
			if doc.Visibility == "public" {
				e.documents[doc.ID] = doc
			}
		}
		return nil
	}

	// 差分記事のみ埋め込みベクトルを計算
	texts := make([]string, len(toProcess))
	for i, doc := range toProcess {
		texts[i] = doc.Title + " " + doc.Content
	}

	// 内部のGetEmbeddingメソッドを呼び出す（外部注入したAPI呼び出し担当、ollama/openAI/deepseek）
	newEmbeddings, err := e.embedder.GetEmbeddings(ctx, texts)
	if err != nil {
		return err
	}

	// 結果を embeddings map にマージ
	for i, doc := range toProcess {
		e.embeddings[doc.ID] = newEmbeddings[i]
		e.documents[doc.ID] = doc
		e.articleUpdated[doc.ID] = time.Now()
	}

	// 未変更のドキュメント情報も最新に更新
	for _, doc := range docs {
		if doc.Visibility == "public" {
			e.documents[doc.ID] = doc
		}
	}

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
	for id, doc := range e.documents {
		emb, ok := e.embeddings[id]
		if !ok {
			continue
		}
		score := cosineSimilarityVec(queryVec, emb)

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
	sort.Slice(results, func(i, j int) bool {
		return results[i].RelevanceScore > results[j].RelevanceScore
	})

	// ⑤ limit で切り詰め
	if limit < len(results) {
		results = results[:limit]
	}

	return results, nil
}
