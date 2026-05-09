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

// OllamaProvider は Ollama ローカル LLM と通信する
type OllamaProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllamaProvider(baseURL string, model string) *OllamaProvider {
	return &OllamaProvider{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

type ollamaGenerateRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
	Format  string                 `json:"format,omitempty"`
}

type ollamaGenerateResponse struct {
	Response string `json:"response"`
}

type ollamaChatRequest struct {
	Model    string                 `json:"model"`
	Messages []chatMessage          `json:"messages"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
	Format   string                 `json:"format,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatResponse struct {
	Message chatMessage `json:"message"`
}

func (p *OllamaProvider) Generate(ctx context.Context, prompt string) (string, error) {
	return p.GenerateWithOptions(ctx, prompt, GenerateOptions{})
}

// GenerateWithOptions はオプションを指定して生成する
func (p *OllamaProvider) GenerateWithOptions(ctx context.Context, prompt string, opts GenerateOptions) (string, error) {
	body := ollamaGenerateRequest{
		Model:  p.model,
		Prompt: prompt,
		Stream: false,
	}

	// オプションが設定されている場合のみ追加
	if opts.Temperature > 0 || opts.Format != "" {
		body.Options = make(map[string]interface{})
		if opts.Temperature > 0 {
			body.Options["temperature"] = opts.Temperature
		}
	}
	if opts.Format != "" {
		body.Format = opts.Format
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/generate", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Error("ollama error", "status", resp.StatusCode, "body", string(respBody))
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var result ollamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result.Response, nil
}

func (p *OllamaProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	return p.ChatWithOptions(ctx, messages, GenerateOptions{})
}

// ChatWithOptions はオプションを指定してチャットする
func (p *OllamaProvider) ChatWithOptions(ctx context.Context, messages []Message, opts GenerateOptions) (string, error) {
	chatMsgs := make([]chatMessage, 0, len(messages))
	for _, m := range messages {
		chatMsgs = append(chatMsgs, chatMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	body := ollamaChatRequest{
		Model:    p.model,
		Messages: chatMsgs,
		Stream:   false,
	}

	// オプションが設定されている場合のみ追加
	if opts.Temperature > 0 || opts.Format != "" {
		body.Options = make(map[string]interface{})
		if opts.Temperature > 0 {
			body.Options["temperature"] = opts.Temperature
		}
	}
	if opts.Format != "" {
		body.Format = opts.Format
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/chat", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama chat request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Error("ollama chat error", "status", resp.StatusCode, "body", string(respBody))
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var result ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result.Message.Content, nil
}
