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

func TestFilterRAGResultsWithLowerThreshold(t *testing.T) {
	// Vector検索の閾値を0.30に下げた後の挙動を確認
	results := []search.SearchResult{
		{ArticleID: "very_low", RelevanceScore: 0.10},  // 0.30未満 → フィルタリング
		{ArticleID: "medium", RelevanceScore: 0.35},    // 0.30以上 → 通過
		{ArticleID: "high", RelevanceScore: 0.72},      // 0.30以上 → 通過
	}

	filtered := filterRAGResults("vector", results)
	if len(filtered) != 2 {
		t.Fatalf("len(filtered) = %d, want 2", len(filtered))
	}

	// 0.10の結果はフィルタリングされていることを確認
	for _, r := range filtered {
		if r.ArticleID == "very_low" {
			t.Fatal("very_low (score=0.10) should be filtered out")
		}
	}
}
