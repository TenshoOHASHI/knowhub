package search

import (
	"context"
	"testing"
)

func TestComputeTermFreq(t *testing.T) {
	tokens := []string{"go", "grpc", "go", "go"}
	freq := computeTermFreq(tokens)

	if freq["go"] != 3 {
		t.Errorf("freq(go) = %d, want 3", freq["go"])
	}
	if freq["grpc"] != 1 {
		t.Errorf("freq(grpc) = %d, want 1", freq["grpc"])
	}
}

func TestComputeBM25IDF(t *testing.T) {
	tokenizedDocs := [][]string{
		{"go", "grpc", "go"},
		{"python", "ai", "python"},
		{"go", "docker"},
	}

	idf := computeBM25IDF(tokenizedDocs)

	// go: 2文書に出現 → log((3 - 2 + 0.5) / (2 + 0.5) + 1) = log(1.5/2.5 + 1) = log(1.6)
	if idf["go"] <= 0 {
		t.Errorf("IDF(go) = %f, want > 0", idf["go"])
	}

	// grpc: 1文書 → レアなので go より IDF が高いべき
	if idf["grpc"] <= idf["go"] {
		t.Errorf("IDF(grpc) = %f should be > IDF(go) = %f", idf["grpc"], idf["go"])
	}
}

func TestBM25Engine_Search(t *testing.T) {
	engine := NewBM25Engine()
	docs := []Document{
		{ID: "1", Title: "Go gRPC", Content: "Go gRPC Go"},
		{ID: "2", Title: "Python AI", Content: "Python AI Python"},
		{ID: "3", Title: "Go Docker", Content: "Go Docker"},
	}

	err := engine.Index(context.Background(), docs)
	if err != nil {
		t.Fatalf("Index failed: %v", err)
	}

	// avgDl が正しく計算されているか
	if engine.avgDl == 0 {
		t.Error("avgDl should not be 0")
	}

	t.Run("search gRPC returns doc1 first", func(t *testing.T) {
		results, err := engine.Search(context.Background(), "gRPC", 10)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("expected results, got empty")
		}
		if results[0].ArticleID != "1" {
			t.Errorf("top result ID = %s, want 1", results[0].ArticleID)
		}
	})

	t.Run("search Python returns doc2 first", func(t *testing.T) {
		results, err := engine.Search(context.Background(), "Python", 10)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(results) == 0 {
			t.Fatal("expected results, got empty")
		}
		if results[0].ArticleID != "2" {
			t.Errorf("top result ID = %s, want 2", results[0].ArticleID)
		}
	})

	t.Run("search Go returns results with doc containing more go higher", func(t *testing.T) {
		results, err := engine.Search(context.Background(), "Go", 10)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		// "go" を含むのは文書1（3回）と文書3（1回）
		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}
		// 文書1のほうが "go" の出現回数が多い → スコアが高い
		if results[0].ArticleID != "1" {
			t.Errorf("top result ID = %s, want 1 (more go occurrences)", results[0].ArticleID)
		}
	})

	t.Run("search unknown returns empty", func(t *testing.T) {
		results, err := engine.Search(context.Background(), "Rust", 10)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	})
}
