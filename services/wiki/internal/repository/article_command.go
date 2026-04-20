package repository

import (
	"context"
	"database/sql"

	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/model"
)

// CommandRepository はWrite操作のみ
type ArticleCommandRepository interface {
	Create(ctx context.Context, article *model.Article) error
	Save(ctx context.Context, article *model.Article) error
	Delete(ctx context.Context, id string) error
}

type mysqlCommandRepository struct {
	db *sql.DB
}

func NewMysqlCommandRepository(db *sql.DB) ArticleCommandRepository {
	return &mysqlCommandRepository{db: db}
}

func (r *mysqlCommandRepository) Create(ctx context.Context, article *model.Article) error {
	// プレスホルダー(SQLインジェクション対策)
	query := `INSERT INTO articles (id, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		article.ID,
		article.Title,
		article.Content,
		article.CreatedAt,
		article.UpdatedAt,
	)
	return err
}

func (r *mysqlCommandRepository) Save(ctx context.Context, article *model.Article) error {
	// プレスホルダー
	query := `UPDATE articles SET title=?, content=?, updated_at=? WHERE id=?`

	_, err := r.db.ExecContext(ctx, query,
		article.Title,
		article.Content,
		article.UpdatedAt,
		article.ID,
	)
	return err
}

func (r *mysqlCommandRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM articles where id=?`
	// INSERT/UPDATE/DELETEは０件更新でも sql.ErrNoRows にならない
	_, err := r.db.ExecContext(ctx, query, id)

	return err
}
