package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

func NewUser(username, email, password string) (*User, error) {
	// バリデーション
	if username == "" || email == "" || password == "" {
		return nil, fmt.Errorf("username or email or password is required")
	}
	// パスワードをハッシュ化
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password")
	}

	// UUID
	uid := uuid.New().String()

	// Userを返す
	return &User{
		ID:           uid,
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}, nil
}
