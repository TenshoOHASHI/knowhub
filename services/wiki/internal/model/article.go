package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// フィール名は大文字（公開）
type Article struct {
	ID         string
	Title      string
	Content    string
	CategoryID string
	Visibility string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewArticle(title string, content string, categoryID string, visibility string) (*Article, error) {
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}
	if visibility == "" || (visibility != "public" && visibility != "locked") {
		visibility = "public"
	}
	uid := uuid.New().String()
	createAt := time.Now()

	return &Article{
		ID:         uid,
		Title:      title,
		Content:    content,
		CategoryID: categoryID,
		Visibility: visibility,
		CreatedAt:  createAt,
		UpdatedAt:  createAt,
	}, nil
}

// 元の記事を直接変更（インスタンスを生成する必要がない）
func (a *Article) Update(title string, content string, visibility string) {
	if title != "" {
		a.Title = title // 既存の値を上書き
	}
	if content != "" {
		a.Content = content
	}
	if visibility != "" {
		a.Visibility = visibility
	}
	a.UpdatedAt = time.Now()
}
