package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// DeepSeekProvider は DeepSeek API と通信する
// OpenAI 互換フォーマットを使用
type DeepSeekProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewDeepSeekProvider(apiKey, model string) *DeepSeekProvider {
	return &DeepSeekProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

func (p *DeepSeekProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.Chat(ctx, []Message{
		{Role: "user", Content: prompt},
	})
}

func (p *DeepSeekProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	msgs := make([]chatMessageGLM5, 0, len(messages))
	for _, m := range messages {
		msgs = append(msgs, chatMessageGLM5{Role: m.Role, Content: m.Content})
	}

	body := chatCompletionRequest{
		Model:    p.model,
		Messages: msgs,
		Stream:   false,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.deepseek.com/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("DeepSeek request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Error("DeepSeek error", "status", resp.StatusCode, "body", string(respBody))
		return "", fmt.Errorf("DeepSeek returned status %d", resp.StatusCode)
	}

	var result chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("DeepSeek returned no choices")
	}

	return result.Choices[0].Message.Content, nil
}
