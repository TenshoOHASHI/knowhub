package embedding

import "context"

type EmbeddingProvider interface {
	// クエリ１件のテキストをベクトルに変換
	GetEmbedding(ctx context.Context, text string) ([]float64, error)
	// Index 時に全記事を一括変するために使う（複数のデータを返す）
	GetEmbeddings(ctx context.Context, text []string) ([][]float64, error)
}
