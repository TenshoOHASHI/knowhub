package search

import (
	"context"
	"strings"
	"testing"

	"github.com/TenshoOHASHI/knowhub/services/ai/internal/llm"
)

type fakeEmbeddingProvider struct{}

func (p fakeEmbeddingProvider) GetEmbedding(ctx context.Context, text string) ([]float64, error) {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "remote procedure") || strings.Contains(lower, "semantic"):
		return []float64{1, 0}, nil
	case strings.Contains(lower, "grpc"):
		return []float64{0.6, 0.8}, nil
	case strings.Contains(lower, "python"):
		return []float64{0, 1}, nil
	default:
		return []float64{0, 0}, nil
	}
}

func (p fakeEmbeddingProvider) GetEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	embeddings := make([][]float64, len(texts))
	for i, text := range texts {
		embedding, err := p.GetEmbedding(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = embedding
	}
	return embeddings, nil
}

type fakeLLMProvider struct{}

func (p fakeLLMProvider) Generate(ctx context.Context, prompt string) (string, error) {
	switch {
	case strings.Contains(prompt, "gRPC article"):
		return `{
			"entities": [
				{"name": "gRPC", "type": "Protocol"}
			],
			"relations": [
				{"source": "gRPC", "target": "Protocol Buffers", "label": "uses"}
			]
		}`, nil
	case strings.Contains(prompt, "Protocol Buffers article"):
		return `{
			"entities": [
				{"name": "Protocol Buffers", "type": "Technology"}
			],
			"relations": []
		}`, nil
	case strings.Contains(prompt, "JWT article"):
		return `{
			"entities": [
				{"name": "JWT", "type": "Security"}
			],
			"relations": []
		}`, nil
	default:
		return `{
			"entities": [
				{"name": "gRPC", "type": "Protocol"}
			],
			"relations": []
		}`, nil
	}
}

func (p fakeLLMProvider) Chat(ctx context.Context, messages []llm.Message) (string, error) {
	return "", nil
}

func TestCosineSimilarityVec(t *testing.T) {
	if got := cosineSimilarityVec([]float64{1, 0}, []float64{1, 0}); got != 1 {
		t.Fatalf("cosineSimilarityVec(same direction) = %f, want 1", got)
	}
	if got := cosineSimilarityVec([]float64{1, 0}, []float64{0, 1}); got != 0 {
		t.Fatalf("cosineSimilarityVec(orthogonal) = %f, want 0", got)
	}
}

func TestVectorEngine_SearchUsesEmbeddingSimilarity(t *testing.T) {
	engine := NewVectorEngine(fakeEmbeddingProvider{})
	docs := []Document{
		{ID: "1", Title: "Remote Procedure Calls", Content: "Semantic explanation of RPC protocols"},
		{ID: "2", Title: "Python AI", Content: "Python machine learning basics"},
	}

	if err := engine.Index(context.Background(), docs); err != nil {
		t.Fatalf("Index failed: %v", err)
	}

	results, err := engine.Search(context.Background(), "semantic grpc", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected vector search results, got empty")
	}
	if results[0].ArticleID != "1" {
		t.Fatalf("top result ID = %s, want 1", results[0].ArticleID)
	}
}

func TestHybridEngine_SearchMergesBM25AndVectorScores(t *testing.T) {
	engine := NewHybridEngine(fakeEmbeddingProvider{}, 0.2)
	docs := []Document{
		{ID: "1", Title: "gRPC basics", Content: "gRPC server streaming"},
		{ID: "2", Title: "Remote Procedure Calls", Content: "semantic rpc protocol design"},
	}

	if err := engine.Index(context.Background(), docs); err != nil {
		t.Fatalf("Index failed: %v", err)
	}

	results, err := engine.Search(context.Background(), "grpc semantic", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("expected merged hybrid results, got %d", len(results))
	}
	if results[0].ArticleID != "2" {
		t.Fatalf("top result ID = %s, want 2 because vector score is weighted higher", results[0].ArticleID)
	}
}

func TestGraphEngine_SearchFindsRelatedArticlesByBFS(t *testing.T) {
	engine := NewGraphEngine(fakeLLMProvider{})
	docs := []Document{
		{ID: "1", Title: "gRPC article", Content: "gRPC uses HTTP/2 for RPC communication"},
		{ID: "2", Title: "Protocol Buffers article", Content: "Protocol Buffers defines typed messages"},
		{ID: "3", Title: "JWT article", Content: "JWT signs authentication claims"},
	}

	if err := engine.Index(context.Background(), docs); err != nil {
		t.Fatalf("Index failed: %v", err)
	}

	results, err := engine.Search(context.Background(), "gRPC", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	got := make(map[string]bool)
	for _, result := range results {
		got[result.ArticleID] = true
	}
	if !got["1"] {
		t.Fatal("expected direct gRPC article in graph search results")
	}
	if !got["2"] {
		t.Fatal("expected related Protocol Buffers article via graph relation")
	}
	if got["3"] {
		t.Fatal("did not expect unrelated JWT article in graph search results")
	}
}
