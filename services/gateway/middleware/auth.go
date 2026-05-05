package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	pb "github.com/TenshoOHASHI/knowhub/proto/auth"
)

type AuthMiddleWare struct {
	authClient pb.AuthServiceClient
}

func NewAuthMiddleware(client pb.AuthServiceClient) *AuthMiddleWare {
	return &AuthMiddleWare{
		authClient: client,
	}
}

// 公開ルート（認証不要）のホワイトリスト
var publicRouter = map[string]bool{
	"POST /api/user/login":    true,
	"POST /api/user/register": true,
}

func (m *AuthMiddleWare) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path

		// ホワイトリストの場合、次のハンドラーを処理
		if publicRouter[key] {
			next.ServeHTTP(w, r)
			return
		}

		// Swagger UI関連パスは認証不要
		if strings.HasPrefix(r.URL.Path, "/swagger/") {
			next.ServeHTTP(w, r)
			return
		}

		// ai 関連パスは認証不要
		if strings.HasPrefix(r.URL.Path, "/api/ai/") {
			next.ServeHTTP(w, r)
			return
		}

		// analytics 関連パスは認証不要
		if strings.HasPrefix(r.URL.Path, "/api/analytics/") {
			next.ServeHTTP(w, r)
			return
		}

		// like / save は匿名フィンガープリント方式なので認証不要
		if strings.HasPrefix(r.URL.Path, "/api/articles/") &&
			(strings.HasSuffix(r.URL.Path, "/like") || strings.HasSuffix(r.URL.Path, "/save") || strings.HasSuffix(r.URL.Path, "/saved")) {
			next.ServeHTTP(w, r)
			return
		}
		// batch like counts
		if r.URL.Path == "/api/articles/likes" {
			next.ServeHTTP(w, r)
			return
		}

		// Token を Cookie または Authorization ヘッダーから取得
		var tokenStr string
		if cookie, err := r.Cookie("token"); err == nil {
			tokenStr = cookie.Value
		} else if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// GET /api/user/me 以外: token がなくても通す（Optional Auth）
		// token があれば検証して userID を context に保存
		if r.Method == "GET" && r.URL.Path != "/api/user/me" {
			// なければ、通常の処理、userIDは空（非公開情報は返さない）
			if tokenStr != "" {
				// トークンがあれば、認証し、userIDを設定（非公開情報を返す）
				if reps, err := m.authClient.VerifyToken(r.Context(), &pb.VerifyTokenRequest{Token: tokenStr}); err == nil {
					ctx := context.WithValue(r.Context(), "userID", reps.User.Id)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			next.ServeHTTP(w, r)
			return
		}

		if tokenStr == "" {
			http.Error(w, "unauthorized: missing token", http.StatusUnauthorized)
			return
		}

		// verify token
		reps, err := m.authClient.VerifyToken(r.Context(), &pb.VerifyTokenRequest{
			Token: tokenStr,
		})
		if err != nil {
			slog.Error("token verification failed", "error", err)
			http.Error(w, "unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		// save to context with UserID
		ctx := context.WithValue(r.Context(), "userID", reps.User.Id)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
