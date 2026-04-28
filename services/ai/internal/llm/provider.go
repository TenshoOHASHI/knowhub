package llm

import "context"

type LLMProvider interface {
	Generate(ctx context.Context, prompt string) (string, error)

	Chat(ctx context.Context, message []Message) (string, error)
}

type Message struct {
	Role    string //  "system", "user", "assistant"
	Content string
}
