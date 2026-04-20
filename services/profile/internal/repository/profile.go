package repository

import (
	"context"
	"database/sql"

	"github.com/TenshoOHASHI/knowhub/services/profile/internal/model"
)

type ProfileRepository interface {
	// GetProfile
	FindFirst(ctx context.Context) (*model.Profile, error)
	// UpdateProfile
	Save(ctx context.Context, profile *model.Profile) error
}

type mysqlProfileRepository struct {
	db *sql.DB
}

func NewMysqlProfileRepository(db *sql.DB) ProfileRepository {
	return &mysqlProfileRepository{db: db}
}

func (r *mysqlProfileRepository) FindFirst(ctx context.Context) (*model.Profile, error) {
	query := `SELECT id, title, bio, github_url, updated_at From profiles ORDER BY updated_at DESC LIMIT 1`
	row := r.db.QueryRowContext(ctx, query)
	var profile model.Profile

	// フィールドごとにScanする
	err := row.Scan(&profile.ID, &profile.Title, &profile.Bio, &profile.GithubURL, &profile.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &profile, nil
}

func (r *mysqlProfileRepository) Save(ctx context.Context, profile *model.Profile) error {
	query := `UPDATE profiles SET title=?, bio=?, github_url=?, updated_at=? WHERE id=?`
	_, err := r.db.ExecContext(ctx, query,
		profile.Title,
		profile.Bio,
		profile.GithubURL,
		profile.UpdatedAt,
		profile.ID,
	)
	if err != nil {
		return err
	}
	return nil
}
