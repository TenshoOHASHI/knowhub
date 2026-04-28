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

// GLM5Provider は Zhipu AI (GLM-5) API と通信する
// OpenAI 互換フォーマットを使用
type GLM5Provider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewGLM5Provider(apiKey, model string) *GLM5Provider {
	return &GLM5Provider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

// GLM-5 API は OpenAI 互換フォーマット
type chatCompletionRequest struct {
	Model    string            `json:"model"`
	Messages []chatMessageGLM5 `json:"messages"`
	Stream   bool              `json:"stream"`
}

type chatMessageGLM5 struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessageGLM5 `json:"message"`
	} `json:"choices"`
}

func (p *GLM5Provider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.Chat(ctx, []Message{
		{Role: "user", Content: prompt},
	})
}

func (p *GLM5Provider) Chat(ctx context.Context, messages []Message) (string, error) {
	msgs := make([]chatMessageGLM5, 0, len(messages))
	for _, m := range messages {
		msgs = append(msgs, chatMessageGLM5{Role: m.Role, Content: m.Content})
	}

	body := chatCompletionRequest{
		Model:    p.model,
		Messages: msgs,
		Stream:   false,
	}

	// エンコード：Go ->JSON -> utf-8
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://open.bigmodel.cn/api/paas/v4/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("GLM-5 request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Error("GLM-5 error", "status", resp.StatusCode, "body", string(respBody))
		return "", fmt.Errorf("GLM-5 returned status %d", resp.StatusCode)
	}

	var result chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("GLM-5 returned no choices")
	}

	return result.Choices[0].Message.Content, nil
}
