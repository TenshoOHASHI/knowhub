package embedding

import (
	"context"
	"strings"
)

type EmbeddingProvider interface {
	// クエリ１件のテキストをベクトルに変換
	GetEmbedding(ctx context.Context, text string) ([]float64, error)
	// Index 時に全記事を一括変するために使う（複数のデータを返す）
	GetEmbeddings(ctx context.Context, text []string) ([][]float64, error)
}

// NewProvider は apiKey から EmbeddingProvider を自動判定して生成するファクトリ
// apiKey のプレフィックスで判定: 空 → Ollama, それ以外 → apiKey に基づく外部API
func NewProvider(ollamaURL, ollamaModel, apiKey string) EmbeddingProvider {
	// apiKey がなければ Ollama（ローカル）
	if apiKey == "" {
		return NewOllamaProvider(ollamaURL, ollamaModel)
	}
	// apiKey のプレフィックスで判定（LLM と同じプロバイダーのキーを使い回す）
	switch {
	case strings.HasPrefix(apiKey, "sk-") && len(apiKey) > 30:
		// OpenAI キー（sk-... 形式で長い）
		return NewOpenAIProvider(apiKey)
	case strings.Contains(apiKey, "AIza"):
		// Gemini キー（AIza... 形式）
		return NewGeminiProvider(apiKey)
	default:
		// GLM-5 / DeepSeek / その他 → GLM-5 の embedding をデフォルトに
		// （DeepSeek, GLM-5 共に OpenAI 互換フォーマット）
		return NewGLM5Provider(apiKey)
	}
}
