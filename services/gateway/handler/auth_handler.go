package handler

import (
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"

	pb "github.com/TenshoOHASHI/knowhub/proto/auth"
	"github.com/TenshoOHASHI/knowhub/services/gateway/swagger"
	"google.golang.org/grpc"
)

// swagger types (used by swag annotations)
var (
	_ swagger.RegisterRequest
	_ swagger.LoginRequest
)

type AuthHandler struct {
	client               pb.AuthServiceClient
	enablePublicRegister bool
	setupRegisterToken   string
}

func NewAuthHandler(conn *grpc.ClientConn, enablePublicRegister bool, setupRegisterToken string) *AuthHandler {
	return &AuthHandler{
		client:               pb.NewAuthServiceClient(conn),
		enablePublicRegister: enablePublicRegister,
		setupRegisterToken:   setupRegisterToken,
	}
}

// Register ユーザー登録
// @Summary      ユーザー登録
// @Description  新しいユーザーを作成し、JWTをHttpOnly Cookieにセットする
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  swagger.RegisterRequest  true  "登録情報"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {string}  string  "invalid request body"
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/user/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// 本番では誰でも新規ユーザー登録できる状態にしない。
	// ENABLE_PUBLIC_REGISTER=false の場合は、初回セットアップ用tokenを要求する。
	if !h.allowRegister(r) {
		slog.Warn("register request rejected", "remote_addr", r.RemoteAddr)
		http.Error(w, "registration is disabled", http.StatusForbidden)
		return
	}

	// ここでrequiredにする必要ある？
	// あとメールアドレスの型チェックとかも、ドメインでするべきですかね?
	// パスワードの長さとか
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// デコード
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.Register(r.Context(), &pb.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		slog.Error("failed to register", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    resp.Token,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   false, // 本番はhttps通信、true
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":  resp.User,
		"token": resp.Token, // Route Handler 用: body からも token を返す
	})
}

func (h *AuthHandler) allowRegister(r *http.Request) bool {
	// 開発環境向け。trueなら従来通りtokenなしで登録を許可する。
	if h.enablePublicRegister {
		return true
	}
	// 本番でtoken未設定なら、登録APIを完全に閉じる。
	if h.setupRegisterToken == "" {
		return false
	}

	got := r.Header.Get("X-Setup-Token")
	if got == "" {
		return false
	}

	// token比較は処理時間の差から値を推測されにくいConstantTimeCompareを使う。
	return subtle.ConstantTimeCompare([]byte(got), []byte(h.setupRegisterToken)) == 1
}

// Login ログイン
// @Summary      ログイン
// @Description  メールアドレスとパスワードで認証し、JWTをHttpOnly Cookieにセットする
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  swagger.LoginRequest  true  "ログイン情報"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {string}  string  "invalid request body"
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/user/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Decode
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Login
	resp, err := h.client.Login(r.Context(), &pb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		slog.Error("failed to login", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Set Token
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    resp.Token,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   false, // 本番はhttps通信、true
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// TokenはCookieに保存 + Route Handler 用に body にも含める
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":  resp.User,
		"token": resp.Token,
	})

}

// Me ログインユーザー情報取得
// @Summary      ログインユーザー情報
// @Description  Cookie の JWT を検証してユーザー情報を返す
// @Tags         auth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {string}  string  "unauthorized"
// @Router       /api/user/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	// ミドルウエアからuserIDを取り出す
	userID := r.Context().Value("userID")
	if userID == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.client.FindByID(r.Context(), &pb.FindByIDRequest{
		Id: userID.(string),
	})
	if err != nil {
		slog.Error("failed to find user", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": resp.User,
	})
}
