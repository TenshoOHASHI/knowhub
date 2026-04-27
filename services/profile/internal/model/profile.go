package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID          string
	Title       string
	Bio         string
	GithubURL   string
	AvatarURL   string
	TwitterURL  string
	LinkedinURL string
	WantedlyURL string
	Skills      string
	Languages   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewProfile(title, bio, githubUrl, avatarUrl, twitterUrl, linkedinUrl, wantedlyUrl, skills, languages string) (*Profile, error) {

	if title == "" || bio == "" || githubUrl == "" {
		return nil, fmt.Errorf("title, bio and githubUrl are required")
	}

	// UUID
	uid := uuid.New().String()

	return &Profile{
		ID:          uid,
		Title:       title,
		Bio:         bio,
		GithubURL:   githubUrl,
		AvatarURL:   avatarUrl,
		TwitterURL:  twitterUrl,
		LinkedinURL: linkedinUrl,
		WantedlyURL: wantedlyUrl,
		Skills:      skills,
		Languages:   languages,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (p *Profile) Update(title, bio, githubUrl, avatarUrl, twitterUrl, linkedinUrl, wantedlyUrl, skills, languages string) {
	if title != "" {
		p.Title = title
	}
	if bio != "" {
		p.Bio = bio
	}
	if githubUrl != "" {
		p.GithubURL = githubUrl
	}
	p.AvatarURL = avatarUrl
	p.TwitterURL = twitterUrl
	p.LinkedinURL = linkedinUrl
	p.WantedlyURL = wantedlyUrl
	p.Skills = skills
	p.Languages = languages
	p.UpdatedAt = time.Now()
}
