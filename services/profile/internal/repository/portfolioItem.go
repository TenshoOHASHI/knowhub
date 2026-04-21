package repository

import (
	"context"
	"database/sql"

	"github.com/TenshoOHASHI/knowhub/services/profile/internal/model"
)

type PortfolioItemRepository interface {

	// CreatePortfolioItem
	Create(ctx context.Context, portfolioItem *model.PortfolioItem) error
	// ListPortfolioItems
	FindAll(ctx context.Context) ([]*model.PortfolioItem, error)
	// GetPortfolioItems
	FindById(ctx context.Context, id string) (*model.PortfolioItem, error)
	// UpdatePortfolioItem
	Save(ctx context.Context, portfolioItem *model.PortfolioItem) error
	// DeletePortfolioItem
	Delete(ctx context.Context, id string) error
	// GetPortfolioItem

}

// portfolioItem.go
type mysqlPortfolioItemRepository struct {
	db *sql.DB
}

func NewMysqlPortfolioItemRepository(db *sql.DB) PortfolioItemRepository {
	return &mysqlPortfolioItemRepository{db: db}
}

func (r *mysqlPortfolioItemRepository) Create(ctx context.Context, portfolioItem *model.PortfolioItem) error {
	query := `INSERT INTO portfolio_items (id, title, description, url, status, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		portfolioItem.ID,
		portfolioItem.Title,
		portfolioItem.Description,
		portfolioItem.URL,
		portfolioItem.Status,
		portfolioItem.CreatedAt,
	)

	if err != nil {
		return err
	}
	return nil

}

func (r *mysqlPortfolioItemRepository) FindAll(ctx context.Context) ([]*model.PortfolioItem, error) {
	query := `SELECT id, title, description, url, status, created_at FROM portfolio_items`

	rows, err := r.db.QueryContext(ctx, query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var portfolioItems []*model.PortfolioItem
	for rows.Next() {
		var portfolioItem model.PortfolioItem
		rows.Scan(&portfolioItem.ID, &portfolioItem.Title, &portfolioItem.Description, &portfolioItem.URL, &portfolioItem.Status, &portfolioItem.CreatedAt)
		portfolioItems = append(portfolioItems, &portfolioItem)
	}

	return portfolioItems, nil
}

func (r *mysqlPortfolioItemRepository) FindById(ctx context.Context, id string) (*model.PortfolioItem, error) {
	query := `SELECT id, title, description, url, status, created_at FROM portfolio_items WHERE id=?`

	row := r.db.QueryRowContext(ctx, query, id)

	var item model.PortfolioItem
	err := row.Scan(&item.ID, &item.Title, &item.Description, &item.URL, &item.Status, &item.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &item, nil
}
func (r *mysqlPortfolioItemRepository) Save(ctx context.Context, portfolioItem *model.PortfolioItem) error {
	query := `UPDATE portfolio_items SET title=?, description=?, url=?, status=? WHERE id=?`
	_, err := r.db.ExecContext(ctx, query, portfolioItem.Title, portfolioItem.Description, portfolioItem.URL, portfolioItem.Status, portfolioItem.ID)
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
