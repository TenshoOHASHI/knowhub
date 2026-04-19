package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// フィール名は大文字（公開）
type Article struct {
	ID        string
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewArticle(title string, content string) (*Article, error) {
	// Validation
	// const title = 1 -> ドメインではなく、ハンドラーの部分ですよね、ここではしない？
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}
	// Create UUID
	uid := uuid.New().String()

	// Set Date
	createAt := time.Now()
	// Return new instance of article
	return &Article{
		ID:        uid,
		Title:     title,
		Content:   content,
		CreatedAt: createAt,
		UpdatedAt: createAt,
	}, nil
}

// 元の記事を直接変更（インスタンスを生成する必要がない）
func (a *Article) Update(title string, content string) {
	if title != "" {
		a.Title = title // 既存の値を上書き
	}
	if content != "" {
		a.Content = content
	}
	a.UpdatedAt = time.Now()
}
