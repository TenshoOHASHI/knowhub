package repository

import (
	"context"
	"database/sql"

	"github.com/TenshoOHASHI/knowhub/pkg/dbutil"
	"github.com/TenshoOHASHI/knowhub/services/profile/internal/model"
)

type PortfolioItemRepository interface {
	Create(ctx context.Context, portfolioItem *model.PortfolioItem) error
	FindAll(ctx context.Context) ([]*model.PortfolioItem, error)
	FindById(ctx context.Context, id string) (*model.PortfolioItem, error)
	Save(ctx context.Context, portfolioItem *model.PortfolioItem) error
	Delete(ctx context.Context, id string) error
}

type mysqlPortfolioItemRepository struct {
	db dbutil.DB
}

func NewMysqlPortfolioItemRepository(db dbutil.DB) PortfolioItemRepository {
	return &mysqlPortfolioItemRepository{db: db}
}

func (r *mysqlPortfolioItemRepository) Create(ctx context.Context, portfolioItem *model.PortfolioItem) error {
	query := `INSERT INTO portfolio_items (id, title, description, url, status, category, tech_stack, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		portfolioItem.ID,
		portfolioItem.Title,
		portfolioItem.Description,
		portfolioItem.URL,
		portfolioItem.Status,
		portfolioItem.Category,
		portfolioItem.TechStack,
		portfolioItem.CreatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *mysqlPortfolioItemRepository) FindAll(ctx context.Context) ([]*model.PortfolioItem, error) {
	query := `SELECT id, title, description, url, status, category, tech_stack, created_at FROM portfolio_items`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var portfolioItems []*model.PortfolioItem
	for rows.Next() {
		var item model.PortfolioItem
		rows.Scan(&item.ID, &item.Title, &item.Description, &item.URL, &item.Status, &item.Category, &item.TechStack, &item.CreatedAt)
		portfolioItems = append(portfolioItems, &item)
	}
	return portfolioItems, nil
}

func (r *mysqlPortfolioItemRepository) FindById(ctx context.Context, id string) (*model.PortfolioItem, error) {
	query := `SELECT id, title, description, url, status, category, tech_stack, created_at FROM portfolio_items WHERE id=?`
	row := r.db.QueryRowContext(ctx, query, id)
	var item model.PortfolioItem
	err := row.Scan(&item.ID, &item.Title, &item.Description, &item.URL, &item.Status, &item.Category, &item.TechStack, &item.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &item, nil
}

func (r *mysqlPortfolioItemRepository) Save(ctx context.Context, portfolioItem *model.PortfolioItem) error {
	query := `UPDATE portfolio_items SET title=?, description=?, url=?, status=?, category=?, tech_stack=? WHERE id=?`
	_, err := r.db.ExecContext(ctx, query, portfolioItem.Title, portfolioItem.Description, portfolioItem.URL, portfolioItem.Status, portfolioItem.Category, portfolioItem.TechStack, portfolioItem.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *mysqlPortfolioItemRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM portfolio_items WHERE id=?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}
