package repository

import (
	"context"
	"database/sql"

	"github.com/TenshoOHASHI/knowhub/pkg/dbutil"
	"github.com/TenshoOHASHI/knowhub/services/auth/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
}

type mysqlRepository struct {
	db dbutil.DB
}

func NewMysqlRepository(db dbutil.DB) UserRepository {
	return &mysqlRepository{
		db: db,
	}
}

func (r *mysqlRepository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (id, username, email, password_hash, created_at) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
	)

	if err != nil {
		return err
	}
	return nil
}

func (r *mysqlRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, username, email, password_hash, created_at From users WHERE email=?`
	row := r.db.QueryRowContext(ctx, query, email)

	var user model.User
	// ポインターに直接値を書き込む
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *mysqlRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE id=? `
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &user, nil
}
