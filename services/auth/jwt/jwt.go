package jwt

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID               string `json:"user_id"`
	Username             string `json:"username"`
	jwt.RegisteredClaims        // expを埋め込む
}

func GenerateToken(userID, username string) (string, error) {
	// 秘密鍵を読み込む
	keyDate, err := os.ReadFile("keys/private.pem")
	if err != nil {
		return "", fmt.Errorf("failed to read private key: %w", err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyDate)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %w", err)
	}

	// ペイロードを作成
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24時間有効
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

func VerifyToken(tokenString string) (*Claims, error) {
	keyData, err := os.ReadFile("keys/public.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse publicKey key: %w", err)
	}

	// トークンを検証、クレームを取り出す
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
