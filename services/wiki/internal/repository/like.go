package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/TenshoOHASHI/knowhub/pkg/dbutil"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type LikeRepository interface {
	ToggleLike(ctx context.Context, articleID, fingerprint string) (count int32, liked bool, err error)
	GetLikeCount(ctx context.Context, articleID, fingerprint string) (count int32, liked bool, err error)
	GetLikeCounts(ctx context.Context, articleIDs []string) (map[string]int32, error)
}

type mysqlLikeRepository struct {
	db dbutil.DB
}

func NewMysqlLikeRepository(db dbutil.DB) LikeRepository {
	return &mysqlLikeRepository{db: db}
}

func (r *mysqlLikeRepository) ToggleLike(ctx context.Context, articleID, fingerprint string) (int32, bool, error) {
	// Try to delete existing like
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM article_likes WHERE article_id = ? AND fingerprint = ?",
		articleID, fingerprint,
	)
	if err != nil {
		return 0, false, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		// Unliked — get remaining count
		count, err := r.getCount(ctx, articleID)
		return count, false, err
	}

	// No existing like — insert new
	id := uuid.New().String()
	_, err = r.db.ExecContext(ctx,
		"INSERT INTO article_likes (id, article_id, fingerprint, created_at) VALUES (?, ?, ?, ?)",
		id, articleID, fingerprint, time.Now(),
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			count, countErr := r.getCount(ctx, articleID)
			return count, true, countErr
		}
		return 0, false, err
	}

	count, err := r.getCount(ctx, articleID)
	return count, true, err
}

func (r *mysqlLikeRepository) GetLikeCount(ctx context.Context, articleID, fingerprint string) (int32, bool, error) {
	count, err := r.getCount(ctx, articleID)
	if err != nil {
		return 0, false, err
	}

	var exists int
	err = r.db.QueryRowContext(ctx,
		"SELECT 1 FROM article_likes WHERE article_id = ? AND fingerprint = ? LIMIT 1",
		articleID, fingerprint,
	).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return 0, false, err
	}
	liked := err == nil

	return count, liked, nil
}

func (r *mysqlLikeRepository) GetLikeCounts(ctx context.Context, articleIDs []string) (map[string]int32, error) {
	result := make(map[string]int32)
	if len(articleIDs) == 0 {
		return result, nil
	}

	// Build IN clause
	query := "SELECT article_id, COUNT(*) as cnt FROM article_likes WHERE article_id IN ("
	args := make([]interface{}, len(articleIDs))
	for i, id := range articleIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ") GROUP BY article_id"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var articleID string
		var cnt int32
		if err := rows.Scan(&articleID, &cnt); err != nil {
			return nil, err
		}
		result[articleID] = cnt
	}

	return result, nil
}

func (r *mysqlLikeRepository) getCount(ctx context.Context, articleID string) (int32, error) {
	var count int32
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM article_likes WHERE article_id = ?",
		articleID,
	).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return count, nil
}
