package model_test

import (
	"testing"

	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/model"
)

func TestNewArticle_Success(t *testing.T) {
	article, err := model.NewArticle("Go入門", "gRPCとは...", "123", "public")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if article.Title != "Go入門" {
		t.Errorf("expected title Go入門, got %s", article.Title)
	}

	if article.Content != "gRPCとは..." {
		t.Errorf("expected content gRPCとは..., got %s", article.Content)
	}

	if article.ID == "" {
		t.Error("expected UD to be set")
	}

	if article.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if article.UpdatedAt.IsZero() {
		t.Error("expected to be set")
	}

	if !article.CreatedAt.Equal(article.UpdatedAt) {
		t.Error("expected CreatedAt Equal UpdatedAt for new article")
	}
}

func TestNewArticle_Empty(t *testing.T) {
	_, err := model.NewArticle("", "content", "123", "public")

	//　nilの場合
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// タイトルは必須
	if err.Error() != "title is required" {
		t.Errorf("expected 'title is required', got %s", err.Error())
	}
}

func TestNewArticle_EmptyContent(t *testing.T) {
	_, err := model.NewArticle("title", "", "123", "public")

	//　nilの場合
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// タイトルは必須
	if err.Error() != "content is required" {
		t.Errorf("expected 'content is required', got %s", err.Error())
	}

}

func TestNewArticle_DefaultVisibility(t *testing.T) {
	article, err := model.NewArticle("title", "content", "123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if article.Visibility != "public" {
		t.Errorf("expected visibility public, got %s", article.Visibility)
	}

	article2, err := model.NewArticle("title", "content", "123", "invalid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if article2.Visibility != "public" {
		t.Errorf("expected visibility public for invalid value, got %s", article2.Visibility)
	}
}

func TestNewArticle_LockedVisibility(t *testing.T) {
	article, err := model.NewArticle("title", "content", "123", "locked")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if article.Visibility != "locked" {
		t.Errorf("expected visibility locked, got %s", article.Visibility)
	}
}

func TestUpdate(t *testing.T) {
	// Title, Content, UpdatedAt
	article, _ := model.NewArticle("Go入門", "content", "123", "public")
	originalContent := article.Content
	createAt := article.CreatedAt

	article.Update("Go中級", "", "")

	if article.Title != "Go中級" {
		t.Errorf("expected title Go中級, got %v", article.Title)
	}

	if article.Content != originalContent {
		t.Error("content should not change")
	}

	if !article.CreatedAt.Equal(createAt) {
		t.Error("CreatedAt should not change")
	}

	// 作成時刻と違っていれば、更新時刻が更新されている
	if article.UpdatedAt.Equal(createAt) {
		t.Error("UpdatedAt should not change")
	}
}

func TestUpdate_Visibility(t *testing.T) {
	article, _ := model.NewArticle("Go入門", "content", "123", "public")

	article.Update("", "", "locked")

	if article.Visibility != "locked" {
		t.Errorf("expected visibility locked, got %s", article.Visibility)
	}

	// 空文字なら変更されない
	article.Update("", "", "")
	if article.Visibility != "locked" {
		t.Errorf("expected visibility locked (unchanged), got %s", article.Visibility)
	}
}
