package search

import (
	"context"
	"math"
	"testing"
)

func TestTokenize(t *testing.T) {
	// 日本語の文字種境界で分割できるか
	got := tokenize("Go言語でgRPC")
	want := []string{"go", "言語", "で", "grpc"}

	if len(got) != len(want) {
		t.Fatalf("tokenize = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("tokenize[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestComputeTF(t *testing.T) {
	tf := computeTF([]string{"go", "grpc", "go"})

	// go が 2/3 ≈ 0.667
	if math.Abs(tf["go"]-2.0/3.0) > 1e-9 {
		t.Errorf("TF(go) = %f, want %f", tf["go"], 2.0/3.0)
	}
}

func TestComputeIDF(t *testing.T) {
	idf := computeIDF([][]string{
		{"go", "grpc", "go"},
		{"python", "ai"},
		{"go", "docker"},
	})

	// grpc(1文書のみ) の IDF > go(2文書) の IDF
	if idf["grpc"] <= idf["go"] {
		t.Errorf("レア単語(grpc)のIDF = %f が頻出単語(go)のIDF = %f 以下", idf["grpc"], idf["go"])
	}
}

func TestCosineSimilarity(t *testing.T) {
	a := []float64{1.0, 2.0, 3.0}
	score := cosineSimilarity(a, a)

	// 同じベクトル同士 → 1.0
	if math.Abs(score-1.0) > 1e-9 {
		t.Errorf("cos(a,a) = %f, want 1.0", score)
	}
}

func TestTFIDFEngine_Search(t *testing.T) {
	engine := NewTFIDFEngine()
	engine.Index(context.Background(), []Document{
		{ID: "1", Title: "Go gRPC", Content: "Go gRPC Go"},
		{ID: "2", Title: "Python AI", Content: "Python AI Python"},
	})

	results, _ := engine.Search(context.Background(), "gRPC", 10)

	if len(results) == 0 {
		t.Fatal("結果が空")
	}
	if results[0].ArticleID != "1" {
		t.Errorf("1位のID = %s, want 1", results[0].ArticleID)
	}
}
