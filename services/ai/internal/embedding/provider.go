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

// NewProvider は model + apiKey から EmbeddingProvider を自動判定して生成するファクトリ
// model プレフィックスで判定: llm.NewProvider と同じロジック
func NewProvider(ollamaURL, ollamaModel, apiKey, model string) EmbeddingProvider {
	// apiKey がなければ Ollama（ローカル）
	if apiKey == "" {
		return NewOllamaProvider(ollamaURL, ollamaModel)
	}
	// model プレフィックスで判定（llm.NewProvider と整合）
	switch {
	case strings.HasPrefix(model, "deepseek"):
		// DeepSeek は embedding API を提供していない → Ollama ローカルにフォールバック
		return NewOllamaProvider(ollamaURL, ollamaModel)
	case strings.HasPrefix(model, "gemini"):
		return NewGeminiProvider(apiKey)
	case strings.HasPrefix(model, "glm"):
		return NewGLM5Provider(apiKey)
	case strings.HasPrefix(model, "gpt"):
		return NewOpenAIProvider(apiKey)
	default:
		// Ollama 等のローカルモデル
		return NewOllamaProvider(ollamaURL, ollamaModel)
	}
}
