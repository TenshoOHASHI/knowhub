package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OllamaEmbeddingProvider struct {
	baseURL string
	model   string // "nomic-embed-text"
}

func NewOllamaEmbeddingProvider(baseURL, model string) *OllamaEmbeddingProvider {
	return &OllamaEmbeddingProvider{
		baseURL: baseURL,
		model:   model,
	}
}

type ollamaEmbedRequest struct {
	Model string      `json:"model"`
	Input interface{} `json:"input"` // string or []string 両対応
}

// ポイント2: レスポンス構造体
type ollamaEmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

func (p *OllamaEmbeddingProvider) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	emeds, err := p.GetEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}

	if len(emeds) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return emeds[0], nil
}

func (p *OllamaEmbeddingProvider) GetEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	resBody := ollamaEmbedRequest{
		Model: p.model,
		Input: texts,
	}

	// json列に変換
	body, err := json.Marshal(resBody)
	if err != nil {
		return nil, err
	}

	// リクエストを作成
	url := p.baseURL + "/api/embed"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))

	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// リクエストを実行
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
		return nil, fmt.Errorf("ollama embed failed: %s", string(respBody))
	}

	var result ollamaEmbedResponse

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return result.Embeddings, nil
}
