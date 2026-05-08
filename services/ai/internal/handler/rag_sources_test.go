package handler

import (
	"testing"

	"github.com/TenshoOHASHI/knowhub/services/ai/internal/search"
)

func TestFilterRAGResults(t *testing.T) {
	results := []search.SearchResult{
		{ArticleID: "low", RelevanceScore: 0.10},
		{ArticleID: "high", RelevanceScore: 0.72},
	}

	filtered := filterRAGResults("vector", results)
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d, want 1", len(filtered))
	}
	if filtered[0].ArticleID != "high" {
		t.Fatalf("filtered[0].ArticleID = %s, want high", filtered[0].ArticleID)
	}
}

func TestAnswerIndicatesNoRelevantContext(t *testing.T) {
	if !answerIndicatesNoRelevantContext("提供されたコンテキストには関連情報がありません。参考: ...") {
		t.Fatal("expected no relevant context phrase to be detected")
	}

	if answerIndicatesNoRelevantContext("この記事ではGoのcontextについて説明しています。") {
		t.Fatal("expected related answer not to be detected as unrelated")
	}
}
