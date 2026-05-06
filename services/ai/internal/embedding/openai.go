package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OpenAI 互換 API 共通構造体
// OpenAI / DeepSeek / Gemini / GLM-5 は全て同じ API フォーマットを使う
type openAIEmbeddingProvider struct {
	baseURL string // プロバイダーごとに変わる
	apiKey  string
	model   string
}

// --- リクエスト/レスポンス構造体 ---

type openAIEmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type openAIEmbedResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
}

// --- コンストラクタ（プロバイダーごと） ---

func NewOpenAIProvider(apiKey string) *openAIEmbeddingProvider {
	return &openAIEmbeddingProvider{
		baseURL: "https://api.openai.com/v1",
		apiKey:  apiKey,
		model:   "text-embedding-3-small",
	}
}

func NewDeepSeekProvider(apiKey string) *openAIEmbeddingProvider {
	return &openAIEmbeddingProvider{
		baseURL: "https://api.deepseek.com/v1",
		apiKey:  apiKey,
		model:   "deepseek-chat",
	}
}

func NewGeminiProvider(apiKey string) *openAIEmbeddingProvider {
	return &openAIEmbeddingProvider{
		baseURL: "https://generativelanguage.googleapis.com/v1beta/openai",
		apiKey:  apiKey,
		model:   "gemini-embedding-exp-03-07",
	}
}

func NewGLM5Provider(apiKey string) *openAIEmbeddingProvider {
	return &openAIEmbeddingProvider{
		baseURL: "https://open.bigmodel.cn/api/paas/v4",
		apiKey:  apiKey,
		model:   "embedding-3",
	}
}

// --- GetEmbedding: 1件変換 ---

func (p *openAIEmbeddingProvider) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	embeds, err := p.GetEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeds) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeds[0], nil
}

// --- GetEmbeddings: 複数一括変換 ---

func (p *openAIEmbeddingProvider) GetEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	reqBody := openAIEmbedRequest{
		Model: p.model,
		Input: texts,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := p.baseURL + "/embeddings"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API failed: %s", string(respBody))
	}

	var result openAIEmbedResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	// data[].embedding を [][]float64 に変換
	embeddings := make([][]float64, len(result.Data))
	for i, d := range result.Data {
		embeddings[i] = d.Embedding
	}

	return embeddings, nil
}
