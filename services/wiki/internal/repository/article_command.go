package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/TenshoOHASHI/knowhub/pkg/dbutil"
	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/model"
	"github.com/redis/go-redis/v9"
)

// CommandRepository はWrite操作のみ
type ArticleCommandRepository interface {
	Create(ctx context.Context, article *model.Article) error
	Save(ctx context.Context, article *model.Article) error
	Delete(ctx context.Context, id string) error
}

type mysqlCommandRepository struct {
	db  dbutil.DB
	rdb *redis.Client
}

func NewMysqlCommandRepository(rdb *redis.Client, db dbutil.DB) ArticleCommandRepository {
	return &mysqlCommandRepository{rdb: rdb, db: db}
}

func (r *mysqlCommandRepository) Create(ctx context.Context, article *model.Article) error {
	// プレスホルダー(SQLインジェクション対策)
	query := `INSERT INTO articles (id, title, content, category_id, visibility, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		article.ID,
		article.Title,
		article.Content,
		article.CategoryID,
		article.Visibility,
		article.CreatedAt,
		article.UpdatedAt,
	)
	if err != nil {
		slog.Error("failed to insert article", "error", err)
		return err
	}

	// キャッシュ無効化 — 一覧キャッシュを削除
	r.rdb.Del(ctx, "articles:list")
	return nil
}

func (r *mysqlCommandRepository) Save(ctx context.Context, article *model.Article) error {
	// プレスホルダー
	query := `UPDATE articles SET title=?, content=?, category_id=?, visibility=?, updated_at=? WHERE id=?`

	_, err := r.db.ExecContext(ctx, query,
		article.Title,
		article.Content,
		article.CategoryID,
		article.Visibility,
		article.UpdatedAt,
		article.ID,
	)
	if err != nil {
		slog.Error("failed to update article", "error", err)
	}

	// Save — 個別キャッシュ + 一覧キャッシュを削除
	r.rdb.Del(ctx, fmt.Sprintf("article:%s", article.ID))
	r.rdb.Del(ctx, "articles:list")

	return err
}

func (r *mysqlCommandRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM articles where id=?`
	// INSERT/UPDATE/DELETEは０件更新でも sql.ErrNoRows にならない
	_, err := r.db.ExecContext(ctx, query, id)

	r.rdb.Del(ctx, "articles:list")
	return err
}
