package repository

import (
	"context"

	"github.com/TenshoOHASHI/knowhub/pkg/dbutil"
	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/model"
)

type CategoryRepository interface {
	FindAll(ctx context.Context) ([]*model.Category, error)
	Create(ctx context.Context, category *model.Category) error
	Delete(ctx context.Context, id string) error
}

type mysqlCategoryRepository struct {
	db dbutil.DB
}

func NewMysqlCategoryRepository(db dbutil.DB) CategoryRepository {
	return &mysqlCategoryRepository{db: db}
}

func (r *mysqlCategoryRepository) FindAll(ctx context.Context) ([]*model.Category, error) {
	query := `SELECT id, name, parent_id FROM categories ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.ParentID); err != nil {
			return nil, err
		}
		categories = append(categories, &c)
	}
	return categories, nil
}

func (r *mysqlCategoryRepository) Create(ctx context.Context, category *model.Category) error {
	query := `INSERT INTO categories (id, name, parent_id) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, category.ID, category.Name, category.ParentID)
	return err
}

func (r *mysqlCategoryRepository) Delete(ctx context.Context, id string) error {
	// 紐づく記事のcategory_idをNULLにする
	_, err := r.db.ExecContext(ctx, `UPDATE articles SET category_id = NULL WHERE category_id = ?`, id)
	if err != nil {
		return err
	}

	// 子カテゴリのparent_idを空にする
	_, err = r.db.ExecContext(ctx, `UPDATE categories SET parent_id = '' WHERE parent_id = ?`, id)
	if err != nil {
		return err
	}

	// カテゴリ削除
	_, err = r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = ?`, id)
	return err
}
