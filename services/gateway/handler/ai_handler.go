package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	pb "github.com/TenshoOHASHI/knowhub/proto/ai"

	"google.golang.org/grpc"
)

type AIHandler struct {
	client pb.AIServiceClient
}

func NewAIHandler(conn *grpc.ClientConn) *AIHandler {
	return &AIHandler{
		client: pb.NewAIServiceClient(conn),
	}
}

// SearchArticles — POST /api/ai/search
func (h *AIHandler) SearchArticles(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query string `json:"query"`
		Limit int32  `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.SearchArticles(r.Context(), &pb.SearchRequest{
		Query: req.Query,
		Limit: req.Limit,
	})
	if err != nil {
		slog.Error("failed to search articles", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// SummarizeArticle — POST /api/ai/summarize
func (h *AIHandler) SummarizeArticle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ArticleID string `json:"article_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.SummarizeArticle(r.Context(), &pb.SummarizeRequest{
		ArticleId: req.ArticleID,
	})
	if err != nil {
		slog.Error("failed to summarize article", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AskQuestion — POST /api/ai/ask
func (h *AIHandler) AskQuestion(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Question     string `json:"question"`
		Model        string `json:"model"`
		ApiKey       string `json:"api_key"`
		SearchEngine string `json:"search_engine"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := h.client.AskQuestion(ctx, &pb.QuestionRequest{
		Question:     req.Question,
		Model:        req.Model,
		ApiKey:       req.ApiKey,
		SearchEngine: req.SearchEngine,
	})
	if err != nil {
		slog.Error("failed to ask question", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetKnowledgeGraph — GET /api/ai/graph
func (h *AIHandler) GetKnowledgeGraph(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	resp, err := h.client.GetKnowledgeGraph(ctx, &pb.GetKnowledgeGraphRequest{})
	if err != nil {
		slog.Error("failed to get knowledge graph", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
