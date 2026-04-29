package llm

import (
	"context"
	"log/slog"
	"strings"
)

type LLMProvider interface {
	Generate(ctx context.Context, prompt string) (string, error)

	Chat(ctx context.Context, message []Message) (string, error)
}

type Message struct {
	Role    string //  "system", "user", "assistant"
	Content string
}

func NewProvider(model, apiKey string) LLMProvider {
	slog.Info("req::", model, apiKey)
	switch {
	// 他のモデルに対応するために、prefixを使用。例：deepseek-chat や gemeini-2.0-flash
	case strings.HasPrefix(model, "deepseek"):
		return NewDeepSeekProvider(apiKey, model)
	case strings.HasPrefix(model, "gemini"):
		return NewGeminiProvider(apiKey, model)
	case strings.HasPrefix(model, "glm"):
		return NewGLM5Provider(apiKey, model)
	case strings.HasPrefix(model, "gpt"):
		return NewOpenAIProvider(apiKey)
	default:
		// ollama の場合（apiKey不要）
		return NewOllamaProvider("http://localhost:11434", model)
	}
}
