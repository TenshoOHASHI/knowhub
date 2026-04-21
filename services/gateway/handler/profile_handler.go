package handler

import (
	"encoding/json"
	"net/http"

	pb "github.com/TenshoOHASHI/knowhub/proto/profile"
	"google.golang.org/grpc"
)

type ProfileHandler struct {
	client pb.ProfileServiceClient
}

func NewProfileHandle(conn *grpc.ClientConn) *ProfileHandler {
	return &ProfileHandler{
		client: pb.NewProfileServiceClient(conn),
	}
}

func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.GetProfile(r.Context(), &pb.GetProfileRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *ProfileHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title     string `json:"title"`
		Bio       string `json:"bio"`
		GithubUrl string `json:"github_url"`
	}

	json.NewDecoder(r.Body).Decode(&req)

	resp, err := h.client.CreateProfile(r.Context(), &pb.CreateProfileRequest{
		Title:     req.Title,
		Bio:       req.Bio,
		GithubUrl: req.GithubUrl,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// 自分自身のプロフィールのみ、特定のIDは不要
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {

	var req struct {
		Title     string `json:"title"`
		Bio       string `json:"bio"`
		GithubUrl string `json:"github_url"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request Body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.UpdateProfile(r.Context(), &pb.UpdateProfileRequest{
		Title:     req.Title,
		Bio:       req.Bio,
		GithubUrl: req.GithubUrl,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *ProfileHandler) ListPortfolioItem(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.ListPortfolioItems(r.Context(), &pb.ListPortfolioItemsRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *ProfileHandler) GetPortfolioItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	resp, err := h.client.GetPortfolioItem(r.Context(), &pb.GetPortfolioItemRequest{Id: id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *ProfileHandler) CreatePortfolioItem(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Url         string `json:"url"`
		Status      string `json:"status"`
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
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *ProfileHandler) UpdatePortfolioItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		// 値を指定しない場合は、ゼロ値でnilになる
		Title       *string `json:"title,omitempty"`
		Description *string `json:"description,omitempty"`
		Url         *string `json:"url,omitempty"`
		Status      *string `json:"status,omitempty"`
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
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *ProfileHandler) DeletePortfolioItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	resp, err := h.client.DeletePortfolioItem(r.Context(), &pb.DeletePortfolioItemRequest{
		Id: id,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
