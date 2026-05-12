package repository

import (
	"context"
	"database/sql"

	"github.com/TenshoOHASHI/knowhub/pkg/dbutil"
	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/model"
)

// ここでDBの依存をドメインに向ける（ドメインはインターフェースに依存し、且つインターフェースはドメインに依存）
// モックとしてテストがしやすい（これらのメソッドを持っていれば、自動でArticleのインターフェースを満たす）
type ArticleRepository interface {
	Create(ctx context.Context, article *model.Article) error
	FindById(ctx context.Context, id string) (*model.Article, error)
	FindAll(ctx context.Context) ([]*model.Article, error)
	Save(ctx context.Context, article *model.Article) error
	Delete(ctx context.Context, id string) error
}

// DBの構造体を用意（初期化に使う）
type mysqlRepository struct {
	db dbutil.DB
}

// コンストラクター関数でDBをラップ
func NewMysqlRepository(db dbutil.DB) ArticleRepository {
	return &mysqlRepository{db: db}
}

func (r *mysqlRepository) FindById(ctx context.Context, id string) (*model.Article, error) {
	query := `SELECT id, title, content, category_id, visibility, is_pinned, created_at, updated_at From articles WHERE id=?`

	// 1件取得
	row := r.db.QueryRowContext(ctx, query, id)

	// 型を定義
	var article model.Article
	// DBデータを構造体にマッピング
	err := row.Scan(&article.ID, &article.Title, &article.Content, &article.CategoryID, &article.Visibility, &article.IsPinned, &article.CreatedAt, &article.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// 記事が見つかりません。
			return nil, sql.ErrNoRows
		}
		// 他のエラー
		return nil, err
	}
	return &article, nil
}

func (r *mysqlRepository) FindAll(ctx context.Context) ([]*model.Article, error) {
	query := `SELECT id, title, content, category_id, visibility, is_pinned, created_at, updated_at FROM articles ORDER BY is_pinned DESC, created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close() // 接続が終わったら。プールに戻す

	var articles []*model.Article
	for rows.Next() {
		var article model.Article
		err := rows.Scan(&article.ID, &article.Title, &article.Content, &article.CategoryID, &article.Visibility, &article.IsPinned, &article.CreatedAt, &article.UpdatedAt)
		if err != nil {
			return nil, err
		}
		articles = append(articles, &article)
	}
	return articles, nil
}

func (r *mysqlRepository) Create(ctx context.Context, article *model.Article) error {
	// プレスホルダー(SQLインジェクション対策)
	query := `INSERT INTO articles (id, title, content, category_id, visibility, is_pinned, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		article.ID,
		article.Title,
		article.Content,
		article.CategoryID,
		article.Visibility,
		article.IsPinned,
		article.CreatedAt,
		article.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *mysqlRepository) Save(ctx context.Context, article *model.Article) error {
	// ピン留め設定時、同カテゴリの既存ピン記事を解除
	if article.IsPinned && article.CategoryID != "" {
		unpinQuery := `UPDATE articles SET is_pinned = 0 WHERE category_id = ? AND is_pinned = 1 AND id != ?`
		r.db.ExecContext(ctx, unpinQuery, article.CategoryID, article.ID)
	}

	// プレスホルダー
	query := `UPDATE articles SET title=?, content=?, category_id=?, visibility=?, is_pinned=?, updated_at=? WHERE id=?`

	_, err := r.db.ExecContext(ctx, query,
		article.Title,
		article.Content,
		article.CategoryID,
		article.Visibility,
		article.IsPinned,
		article.UpdatedAt,
		article.ID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *mysqlRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM articles where id=?`
	// INSERT/UPDATE/DELETEは０件更新でも sql.ErrNoRows にならない
	_, err := r.db.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}
	return nil
}
