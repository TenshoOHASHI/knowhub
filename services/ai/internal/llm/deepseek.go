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
	apiKey    string
	model     string
	maxTokens int
	client    *http.Client
}

func NewDeepSeekProvider(apiKey, model string, maxTokens int) *DeepSeekProvider {
	return &DeepSeekProvider{
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
		client:    &http.Client{},
	}
}

func (p *DeepSeekProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.Chat(ctx, []Message{
		{Role: "user", Content: prompt},
	})
}

func (p *DeepSeekProvider) GenerateWithOptions(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	return p.Generate(ctx, prompt)
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
		MaxTokens: p.maxTokens,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	slog.Info("deepseek: sending request", "model", p.model, "messages_count", len(msgs), "body_len", len(jsonBody))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.deepseek.com/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		slog.Error("deepseek: request failed", "error", err)
		return "", fmt.Errorf("DeepSeek request failed: %w", err)
	}
	defer resp.Body.Close()

	slog.Info("deepseek: response received", "status", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Error("DeepSeek error", "status", resp.StatusCode, "body", string(respBody))
		return "", NewHTTPError(resp.StatusCode, string(respBody))
	}

	slog.Info("deepseek: starting to decode response")

	// デコード失敗時にボディをログに出力するため、先に読み取る
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("deepseek: read body failed", "error", err)
		return "", fmt.Errorf("read response body: %w", err)
	}
	slog.Info("deepseek: body read successfully", "body_len", len(respBody))

	var result chatCompletionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		preview := string(respBody)
		if len(preview) > 1000 {
			preview = preview[:1000]
		}
		slog.Error("deepseek: decode failed", "error", err, "body_preview", preview)
		return "", fmt.Errorf("decode response: %w", err)
	}
	slog.Info("deepseek: decoded successfully", "choices", len(result.Choices))

	if len(result.Choices) == 0 {
		slog.Error("deepseek: no choices returned")
		return "", fmt.Errorf("DeepSeek returned no choices")
	}

	slog.Info("deepseek: response content", "content_len", len(result.Choices[0].Message.Content))
	return result.Choices[0].Message.Content, nil
}
