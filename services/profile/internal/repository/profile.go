package repository

import (
	"context"
	"database/sql"

	"github.com/TenshoOHASHI/knowhub/pkg/dbutil"
	"github.com/TenshoOHASHI/knowhub/services/profile/internal/model"
)

type ProfileRepository interface {
	// GetProfile
	FindFirst(ctx context.Context) (*model.Profile, error)
	// UpdateProfile
	Save(ctx context.Context, profile *model.Profile) error
	// CreateProfile
	Create(ctx context.Context, profile *model.Profile) error
}

type mysqlProfileRepository struct {
	db dbutil.DB
}

func NewMysqlProfileRepository(db dbutil.DB) ProfileRepository {
	return &mysqlProfileRepository{db: db}
}

func (r *mysqlProfileRepository) FindFirst(ctx context.Context) (*model.Profile, error) {
	query := `SELECT id, title, bio, github_url, avatar_url, twitter_url, linkedin_url, wantedly_url, skills, languages, created_at, updated_at From profiles ORDER BY updated_at DESC LIMIT 1`
	row := r.db.QueryRowContext(ctx, query)
	var profile model.Profile

	err := row.Scan(&profile.ID, &profile.Title, &profile.Bio, &profile.GithubURL, &profile.AvatarURL, &profile.TwitterURL, &profile.LinkedinURL, &profile.WantedlyURL, &profile.Skills, &profile.Languages, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &profile, nil
}

func (r *mysqlProfileRepository) Save(ctx context.Context, profile *model.Profile) error {
	query := `UPDATE profiles SET title=?, bio=?, github_url=?, avatar_url=?, twitter_url=?, linkedin_url=?, wantedly_url=?, skills=?, languages=?, created_at=?, updated_at=? WHERE id=?`
	_, err := r.db.ExecContext(ctx, query,
		profile.Title,
		profile.Bio,
		profile.GithubURL,
		profile.AvatarURL,
		profile.TwitterURL,
		profile.LinkedinURL,
		profile.WantedlyURL,
		profile.Skills,
		profile.Languages,
		profile.CreatedAt,
		profile.UpdatedAt,
		profile.ID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *mysqlProfileRepository) Create(ctx context.Context, profile *model.Profile) error {
	query := `INSERT INTO profiles (id, title, bio, github_url, avatar_url, twitter_url, linkedin_url, wantedly_url, skills, languages, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, &profile.ID, &profile.Title, &profile.Bio, &profile.GithubURL, &profile.AvatarURL, &profile.TwitterURL, &profile.LinkedinURL, &profile.WantedlyURL, &profile.Skills, &profile.Languages, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}
