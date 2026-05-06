package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/TenshoOHASHI/knowhub/pkg/dbutil"
	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/model"
	"github.com/google/uuid"
)

type SavedArticleRepository interface {
	Save(ctx context.Context, articleID, fingerprint string) error
	Unsave(ctx context.Context, articleID, fingerprint string) error
	IsSaved(ctx context.Context, articleID, fingerprint string) (bool, error)
	ListByFingerprint(ctx context.Context, fingerprint string) ([]*model.Article, error)
}

type mysqlSavedArticleRepository struct {
	db dbutil.DB
}

func NewMysqlSavedArticleRepository(db dbutil.DB) SavedArticleRepository {
	return &mysqlSavedArticleRepository{db: db}
}

func (r *mysqlSavedArticleRepository) Save(ctx context.Context, articleID, fingerprint string) error {
	id := uuid.New().String()
	_, err := r.db.ExecContext(ctx,
		"INSERT IGNORE INTO saved_articles (id, article_id, fingerprint, created_at) VALUES (?, ?, ?, ?)",
		id, articleID, fingerprint, time.Now(),
	)
	return err
}

func (r *mysqlSavedArticleRepository) Unsave(ctx context.Context, articleID, fingerprint string) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM saved_articles WHERE article_id = ? AND fingerprint = ?",
		articleID, fingerprint,
	)
	return err
}

func (r *mysqlSavedArticleRepository) IsSaved(ctx context.Context, articleID, fingerprint string) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx,
		"SELECT 1 FROM saved_articles WHERE article_id = ? AND fingerprint = ? LIMIT 1",
		articleID, fingerprint,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mysqlSavedArticleRepository) ListByFingerprint(ctx context.Context, fingerprint string) ([]*model.Article, error) {
	query := `SELECT a.id, a.title, a.content, a.category_id, a.visibility, a.created_at, a.updated_at
		FROM saved_articles s
		JOIN articles a ON s.article_id = a.id
		WHERE s.fingerprint = ?
		ORDER BY s.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, fingerprint)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []*model.Article
	for rows.Next() {
		var article model.Article
		if err := rows.Scan(&article.ID, &article.Title, &article.Content, &article.CategoryID, &article.Visibility, &article.CreatedAt, &article.UpdatedAt); err != nil {
			return nil, err
		}
		articles = append(articles, &article)
	}
	return articles, nil
}
