package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	pb "github.com/TenshoOHASHI/knowhub/proto/profile"
	"github.com/TenshoOHASHI/knowhub/services/gateway/swagger"
	"google.golang.org/grpc"
)

// swagger types (used by swag annotations)
var (
	_ swagger.CreateProfileRequest
	_ swagger.CreatePortfolioItemRequest
	_ swagger.UpdatePortfolioItemRequest
)

type ProfileHandler struct {
	client pb.ProfileServiceClient
}

func NewProfileHandle(conn *grpc.ClientConn) *ProfileHandler {
	return &ProfileHandler{
		client: pb.NewProfileServiceClient(conn),
	}
}

// GetProfile プロフィール取得
// @Summary      プロフィール取得
// @Description  自身のプロフィールを取得する
// @Tags         profile
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/profile [get]
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.GetProfile(r.Context(), &pb.GetProfileRequest{})
	if err != nil {
		slog.Error("failed to get profile", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateProfile プロフィール作成
// @Summary      プロフィール作成
// @Description  新しいプロフィールを作成する
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        request  body  swagger.CreateProfileRequest  true  "プロフィールデータ"
// @Success      201  {object}  map[string]interface{}
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/profile [post]
func (h *ProfileHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string `json:"title"`
		Bio         string `json:"bio"`
		GithubUrl   string `json:"github_url"`
		AvatarUrl   string `json:"avatar_url"`
		TwitterUrl  string `json:"twitter_url"`
		LinkedinUrl string `json:"linkedin_url"`
		WantedlyUrl string `json:"wantedly_url"`
		Skills      string `json:"skills"`
		Languages   string `json:"languages"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	resp, err := h.client.CreateProfile(r.Context(), &pb.CreateProfileRequest{
		Title:       req.Title,
		Bio:         req.Bio,
		GithubUrl:   req.GithubUrl,
		AvatarUrl:   req.AvatarUrl,
		TwitterUrl:  req.TwitterUrl,
		LinkedinUrl: req.LinkedinUrl,
		WantedlyUrl: req.WantedlyUrl,
		Skills:      req.Skills,
		Languages:   req.Languages,
	})

	if err != nil {
		slog.Error("failed to create profile", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// UpdateProfile プロフィール更新
// @Summary      プロフィール更新
// @Description  自身のプロフィールを更新する（特定のIDは不要）
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        request  body  swagger.CreateProfileRequest  true  "更新データ"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {string}  string  "invalid request body"
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/profile [put]
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {

	var req struct {
		Title       string `json:"title"`
		Bio         string `json:"bio"`
		GithubUrl   string `json:"github_url"`
		AvatarUrl   string `json:"avatar_url"`
		TwitterUrl  string `json:"twitter_url"`
		LinkedinUrl string `json:"linkedin_url"`
		WantedlyUrl string `json:"wantedly_url"`
		Skills      string `json:"skills"`
		Languages   string `json:"languages"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request Body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.UpdateProfile(r.Context(), &pb.UpdateProfileRequest{
		Title:       req.Title,
		Bio:         req.Bio,
		GithubUrl:   req.GithubUrl,
		AvatarUrl:   req.AvatarUrl,
		TwitterUrl:  req.TwitterUrl,
		LinkedinUrl: req.LinkedinUrl,
		WantedlyUrl: req.WantedlyUrl,
		Skills:      req.Skills,
		Languages:   req.Languages,
	})

	if err != nil {
		slog.Error("failed to update profile", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// ListPortfolioItem ポートフォリオ一覧取得
// @Summary      ポートフォリオ一覧取得
// @Description  全てのポートフォリオアイテムを返す
// @Tags         portfolio
// @Produce      json
// @Success      200  {array}   map[string]interface{}
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/portfolio [get]
func (h *ProfileHandler) ListPortfolioItem(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.ListPortfolioItems(r.Context(), &pb.ListPortfolioItemsRequest{})
	if err != nil {
		slog.Error("failed to list portfolio items", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// GetPortfolioItem ポートフォリオ詳細取得
// @Summary      ポートフォリオ詳細取得
// @Description  指定したIDのポートフォリオアイテムを返す
// @Tags         portfolio
// @Produce      json
// @Param        id   path      string  true  "ポートフォリオID"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/portfolio/{id} [get]
func (h *ProfileHandler) GetPortfolioItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	resp, err := h.client.GetPortfolioItem(r.Context(), &pb.GetPortfolioItemRequest{Id: id})
	if err != nil {
		slog.Error("failed to get portfolio item", "error", err, "id", id)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// CreatePortfolioItem ポートフォリオ作成
// @Summary      ポートフォリオ作成
// @Description  新しいポートフォリオアイテムを作成する
// @Tags         portfolio
// @Accept       json
// @Produce      json
// @Param        request  body  swagger.CreatePortfolioItemRequest  true  "ポートフォリオデータ"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {string}  string  "invalid request body"
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/portfolio [post]
func (h *ProfileHandler) CreatePortfolioItem(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Url         string `json:"url"`
		Status      string `json:"status"`
		Category    string `json:"category"`
		TechStack   string `json:"tech_stack"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.CreatePortfolioItem(r.Context(), &pb.CreatePortfolioItemRequest{
		Title:       req.Title,
		Description: req.Description,
		Url:         req.Url,
		Status:      req.Status,
		Category:    req.Category,
		TechStack:   req.TechStack,
	})

	if err != nil {
		slog.Error("failed to create portfolio item", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// UpdatePortfolioItem ポートフォリオ更新
// @Summary      ポートフォリオ更新
// @Description  指定したIDのポートフォリオアイテムを更新する（部分更新対応）
// @Tags         portfolio
// @Accept       json
// @Produce      json
// @Param        id       path      string  true  "ポートフォリオID"
// @Param        request  body      swagger.UpdatePortfolioItemRequest  true  "更新データ"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {string}  string  "invalid request body"
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/portfolio/{id} [put]
func (h *ProfileHandler) UpdatePortfolioItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Title       *string `json:"title,omitempty"`
		Description *string `json:"description,omitempty"`
		Url         *string `json:"url,omitempty"`
		Status      *string `json:"status,omitempty"`
		Category    *string `json:"category,omitempty"`
		TechStack   *string `json:"tech_stack,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.UpdatePortfolioItem(r.Context(), &pb.UpdatePortfolioItemRequest{
		Id:          id,
		Title:       req.Title,
		Description: req.Description,
		Url:         req.Url,
		Status:      req.Status,
		Category:    req.Category,
		TechStack:   req.TechStack,
	})

	if err != nil {
		slog.Error("failed to update portfolio item", "error", err, "id", id)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// DeletePortfolioItem ポートフォリオ削除
// @Summary      ポートフォリオ削除
// @Description  指定したIDのポートフォリオアイテムを削除する
// @Tags         portfolio
// @Param        id   path      string  true  "ポートフォリオID"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/portfolio/{id} [delete]
func (h *ProfileHandler) DeletePortfolioItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	resp, err := h.client.DeletePortfolioItem(r.Context(), &pb.DeletePortfolioItemRequest{
		Id: id,
	})

	if err != nil {
		slog.Error("failed to delete portfolio item", "error", err, "id", id)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
