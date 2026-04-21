package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID        string
	Title     string
	Bio       string
	GithubURL string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewProfile(title, bio, githubUrl string) (*Profile, error) {

	if title == "" || bio == "" || githubUrl == "" {
		return nil, fmt.Errorf("title, bio and githubUrl are required")
	}

	// UUID
	uid := uuid.New().String()

	return &Profile{
		ID:        uid,
		Title:     title,
		Bio:       bio,
		GithubURL: githubUrl,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (p *Profile) Update(title, bio, githubUrl string) {
	if title != "" {
		p.Title = title
	}
	if bio != "" {
		p.Bio = bio
	}

	if githubUrl != "" {
		p.GithubURL = githubUrl
	}
	p.UpdatedAt = time.Now()
}
