package model

import (
	"fmt"

	"github.com/google/uuid"
)

type Category struct {
	ID       string
	Name     string
	ParentID string // 空文字 = ルートカテゴリ
}

func NewCategory(name string, parentID string) (*Category, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	return &Category{
		ID:       uuid.New().String(),
		Name:     name,
		ParentID: parentID,
	}, nil
}
