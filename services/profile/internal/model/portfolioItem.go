package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PortfolioStatus string

const (
	StatusDeveloping PortfolioStatus = "developing"
	StatusCompleted  PortfolioStatus = "completed"
)

type PortfolioItem struct {
	ID          string
	Title       string
	Description string
	URL         string
	Status      PortfolioStatus
	Category    string
	TechStack   string
	CreatedAt   time.Time
}

func NewPortfolioItem(title, description, url string, status PortfolioStatus, category, techStack string) (*PortfolioItem, error) {
	if title == "" || description == "" {
		return nil, fmt.Errorf("title and description are required")
	}

	if status != StatusDeveloping && status != StatusCompleted {
		return nil, fmt.Errorf("status must be developing or completed")
	}
	if category == "" {
		category = "project"
	}
	return &PortfolioItem{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		URL:         url,
		Status:      status,
		Category:    category,
		TechStack:   techStack,
		CreatedAt:   time.Now(),
	}, nil
}

func (i *PortfolioItem) Update(title, description, url string, status PortfolioStatus, category, techStack string) {
	if title != "" {
		i.Title = title
	}
	if description != "" {
		i.Description = description
	}
	if url != "" {
		i.URL = url
	}
	if status == StatusDeveloping || status == StatusCompleted {
		i.Status = status
	}
	if category != "" {
		i.Category = category
	}
	i.TechStack = techStack
}
