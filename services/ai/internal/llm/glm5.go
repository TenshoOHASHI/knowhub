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
	// メッセージ用のバッファを用意
	msgs := make([]chatMessageGLM5, 0, len(messages))
	// 複数のメッセージを取り出し、GLM5用の構造体に変換してから、バッファに追加
	for _, m := range messages {
		msgs = append(msgs, chatMessageGLM5{Role: m.Role, Content: m.Content})
	}

	// 送信用のリクエスト構造体を初期化
	body := chatCompletionRequest{
		Model:    p.model,
		Messages: msgs,
		Stream:   false,
	}

	// エンコード：Goの構造体をJSON文字列（[]byte）に変換（バイト列にシリアライズ）
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	slog.Info("glm5: sending request", "model", p.model, "messages_count", len(msgs), "body_len", len(jsonBody))

	// コンテキスト付きのリクエスト通信を行う
	// 送信先は、openAIと互換がある、url（相手側のサーバ）にデータ送信するための準備をする
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://open.bigmodel.cn/api/paas/v4/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	// リクエスト用のヘッダーを用意
	// また、認証用のBearerを手動でヘッダーに設定（サーバ間の通信はCORSとは関係ない）
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// ここで、実際にデータを送信
	resp, err := p.client.Do(req)
	if err != nil {
		slog.Error("glm5: request failed", "error", err)
		return "", fmt.Errorf("GLM-5 request failed: %w", err)
	}
	defer resp.Body.Close()

	slog.Info("glm5: response received", "status", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		// ディスク（メモリ？）からバイド列を全て読み込む
		respBody, _ := io.ReadAll(resp.Body)
		slog.Error("GLM-5 error", "status", resp.StatusCode, "body", string(respBody))
		return "", fmt.Errorf("GLM-5 returned status %d", resp.StatusCode)
	}

	// 結果用の構造体を用意
	var result chatCompletionResponse
	// 受け取ったバイト列をデコードし、resultにデータを格納（バイト列を読み込む、json文字列に変換、それをgoの構造体に変換（マッピング））
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Error("glm5: decode failed", "error", err)
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		slog.Error("glm5: no choices returned")
		return "", fmt.Errorf("GLM-5 returned no choices")
	}

	slog.Info("glm5: response content", "content_len", len(result.Choices[0].Message.Content))
	return result.Choices[0].Message.Content, nil
}
