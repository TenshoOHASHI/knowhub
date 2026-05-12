package llm

import (
	"context"
	"strings"
)

type LLMProvider interface {
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateWithOptions(ctx context.Context, prompt string, opts GenerateOptions) (string, error)
	Chat(ctx context.Context, message []Message) (string, error)
}

type Message struct {
	Role    string //  "system", "user", "assistant"
	Content string
}

// GenerateOptions は LLM 生成時のオプション（Ollama向け）
type GenerateOptions struct {
	Temperature float64
	Format      string
}

// JSONOptions は構造化出力向けオプション（temperatureはデフォルト値を使用）
var JSONOptions = GenerateOptions{
	Format: "json",
}

func NewProvider(ollamaURL, model, apiKey string, maxTokens int) LLMProvider {

	switch {
	// 他のモデルに対応するために、prefixを使用。例：deepseek-chat や gemeini-2.0-flash
	case strings.HasPrefix(model, "deepseek"):
		return NewDeepSeekProvider(apiKey, model, maxTokens)
	case strings.HasPrefix(model, "gemini"):
		return NewGeminiProvider(apiKey, model)
	case strings.HasPrefix(model, "glm"):
		return NewGLM5Provider(apiKey, model)
	case strings.HasPrefix(model, "gpt"):
		return NewOpenAIProvider(apiKey)
	default:
		// ollama の場合（apiKey不要）
		return NewOllamaProvider(ollamaURL, model)
	}
}
