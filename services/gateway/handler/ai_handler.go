package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	pb "github.com/TenshoOHASHI/knowhub/proto/ai"
	"github.com/TenshoOHASHI/knowhub/services/gateway/middleware"

	"google.golang.org/grpc"
)

// noArticleMessage は記事が見つからなかった場合の固定メッセージのプレフィックス。
// この応答の場合はレート制限カウントをスキップする。
const noArticleMessage = "Wiki内には関連する記事は見つかりませんでした。"

type AIHandler struct {
	client      pb.AIServiceClient
	rateLimiter *middleware.AIRateLimiter
}

func NewAIHandler(conn *grpc.ClientConn, rateLimiter *middleware.AIRateLimiter) *AIHandler {
	return &AIHandler{
		client:      pb.NewAIServiceClient(conn),
		rateLimiter: rateLimiter,
	}
}

// isAnonymous はリクエストが未ログインユーザーかどうかを判定する。
func isAnonymous(r *http.Request) bool {
	userID, ok := r.Context().Value("userID").(string)
	return !ok || userID == ""
}

// isDeepSeekFree はモデルとAPIキーからDeepSeek Free利用かどうかを判定する。
func isDeepSeekFree(model, apiKey string) bool {
	return apiKey == "" && strings.HasPrefix(model, "deepseek")
}

// checkInputLength は質問文字数を検証する（1000文字制限）。
func checkInputLength(w http.ResponseWriter, question string) bool {
	if len([]rune(question)) > 1000 {
		http.Error(w, "質問は1000文字以内にしてください", http.StatusBadRequest)
		return false
	}
	return true
}

// checkRateLimit はリクエスト前のレート制限チェックを行う（カウントなし）。
// モデル種別に応じて専用カウンターをチェックする。
func (h *AIHandler) checkRateLimit(w http.ResponseWriter, r *http.Request, model, apiKey string) bool {
	if !isAnonymous(r) {
		return true
	}
	clientID := middleware.AnonymousClientID(r)
	if isDeepSeekFree(model, apiKey) {
		return h.rateLimiter.CheckDeepSeekDaily(w, clientID)
	}
	// 外部モデル（自分のAPI Key使用）の日次チェック
	return h.rateLimiter.CheckDaily(w, clientID)
}

// incrementRateLimit はレスポンスが成功した場合にレート制限カウンターをインクリメントする。
func (h *AIHandler) incrementRateLimit(r *http.Request, model, apiKey string) {
	if !isAnonymous(r) {
		return
	}
	clientID := middleware.AnonymousClientID(r)
	if isDeepSeekFree(model, apiKey) {
		h.rateLimiter.IncrementDeepSeekDaily(clientID)
	} else {
		h.rateLimiter.IncrementDaily(clientID)
	}
}

// setRateLimitHeaders はレスポンスヘッダーに残り回数を設定する。
func (h *AIHandler) setRateLimitHeaders(w http.ResponseWriter, r *http.Request, model, apiKey string) {
	if !isAnonymous(r) {
		return
	}
	clientID := middleware.AnonymousClientID(r)
	if isDeepSeekFree(model, apiKey) {
		limit := h.rateLimiter.GetDeepSeekDailyLimit()
		remaining := h.rateLimiter.GetRemainingDeepSeekDaily(clientID)
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	} else {
		limit := h.rateLimiter.GetDailyLimit()
		remaining := h.rateLimiter.GetRemainingDaily(clientID)
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	}
}

// rateLimitJSON は残り回数情報をJSON文字列として返す（SSE用）。
func (h *AIHandler) rateLimitJSON(r *http.Request, model, apiKey string) string {
	if !isAnonymous(r) {
		return ""
	}
	clientID := middleware.AnonymousClientID(r)
	var limit, remaining int
	if isDeepSeekFree(model, apiKey) {
		limit = h.rateLimiter.GetDeepSeekDailyLimit()
		remaining = h.rateLimiter.GetRemainingDeepSeekDaily(clientID)
	} else {
		limit = h.rateLimiter.GetDailyLimit()
		remaining = h.rateLimiter.GetRemainingDaily(clientID)
	}
	return fmt.Sprintf(`{"limit":%d,"remaining":%d}`, limit, remaining)
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

	if !checkInputLength(w, req.Question) {
		return
	}

	if !h.checkRateLimit(w, r, req.Model, req.ApiKey) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
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

	// 記事が見つからなかった場合はレート制限をカウントしない
	if !strings.HasPrefix(resp.Answer, noArticleMessage) {
		h.incrementRateLimit(r, req.Model, req.ApiKey)
	}

	h.setRateLimitHeaders(w, r, req.Model, req.ApiKey)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetKnowledgeGraph — GET /api/ai/graph
func (h *AIHandler) GetKnowledgeGraph(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	slog.Info("GetKnowledgeGraph started")

	// Graph RAGの構築には時間がかかるため、タイムアウトを延長（30分）
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Minute)
	defer cancel()

	resp, err := h.client.GetKnowledgeGraph(ctx, &pb.GetKnowledgeGraphRequest{})
	if err != nil {
		slog.Error("failed to get knowledge graph", "error", err, "duration", time.Since(start))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	slog.Info("GetKnowledgeGraph completed", "duration", time.Since(start), "entities", len(resp.Entities))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetRelatedArticles — GET /api/ai/articles/:id/related
func (h *AIHandler) GetRelatedArticles(w http.ResponseWriter, r *http.Request) {
	articleID := r.PathValue("id")
	if articleID == "" {
		http.Error(w, "article_id is required", http.StatusBadRequest)
		return
	}

	// クエリパラメータ
	maxHops := int32(2)
	limit := int32(10)

	if mh := r.URL.Query().Get("max_hops"); mh != "" {
		fmt.Sscanf(mh, "%d", &maxHops)
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	resp, err := h.client.GetRelatedArticles(ctx, &pb.GetRelatedArticlesRequest{
		ArticleId: articleID,
		MaxHops:   maxHops,
		Limit:     limit,
	})
	if err != nil {
		slog.Error("failed to get related articles", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": resp.Results,
	})
}

// AskWithAgent — POST /api/ai/agent
func (h *AIHandler) AskWithAgent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Question        string `json:"question"`
		Model           string `json:"model"`
		ApiKey          string `json:"api_key"`
		SearchEngine    string `json:"search_engine"`
		EnableWebSearch bool   `json:"enable_web_search"`
		History         string `json:"history"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if !checkInputLength(w, req.Question) {
		return
	}

	if !h.checkRateLimit(w, r, req.Model, req.ApiKey) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	resp, err := h.client.AskWithAgent(ctx, &pb.AgentQuestionRequest{
		Question:        req.Question,
		Model:           req.Model,
		ApiKey:          req.ApiKey,
		SearchEngine:    req.SearchEngine,
		EnableWebSearch: req.EnableWebSearch,
		History:         req.History,
	})
	if err != nil {
		slog.Error("failed to ask with agent", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// 記事が見つからなかった場合はレート制限をカウントしない
	if !strings.HasPrefix(resp.Answer, noArticleMessage) {
		h.incrementRateLimit(r, req.Model, req.ApiKey)
	}

	h.setRateLimitHeaders(w, r, req.Model, req.ApiKey)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AskWithAgentStream — POST /api/ai/agent/stream (SSE)
func (h *AIHandler) AskWithAgentStream(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Question        string `json:"question"`
		Model           string `json:"model"`
		ApiKey          string `json:"api_key"`
		SearchEngine    string `json:"search_engine"`
		EnableWebSearch bool   `json:"enable_web_search"`
		History         string `json:"history"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if !checkInputLength(w, req.Question) {
		return
	}

	if !h.checkRateLimit(w, r, req.Model, req.ApiKey) {
		return
	}

	// SSE ヘッダー設定
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // nginx バッファリング無効化

	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Second)
	defer cancel()

	// gRPC サーバーストリーミング開始
	stream, err := h.client.AskWithAgentStream(ctx, &pb.AgentQuestionRequest{
		Question:        req.Question,
		Model:           req.Model,
		ApiKey:          req.ApiKey,
		SearchEngine:    req.SearchEngine,
		EnableWebSearch: req.EnableWebSearch,
		History:         req.History,
	})
	if err != nil {
		slog.Error("failed to start agent stream", "error", err)
		writeSSE(w, "error", fmt.Sprintf(`{"message":"%s"}`, jsonEscape(err.Error())))
		w.(http.Flusher).Flush()
		return
	}

	flusher, canFlush := w.(http.Flusher)

	var finalAnswer string
	for {
		event, err := stream.Recv()
		if err != nil {
			break
		}

		// final_answer イベントから回答を取得
		if event.EventType == "final_answer" && event.Final != nil {
			finalAnswer = event.Final.Answer
		}

		data, err := json.Marshal(event)
		if err != nil {
			continue
		}

		writeSSE(w, event.EventType, string(data))

		if canFlush {
			flusher.Flush()
		}
	}

	// ストリーム完了後、記事が見つかった場合のみレート制限をカウント
	if finalAnswer != "" && !strings.HasPrefix(finalAnswer, noArticleMessage) {
		h.incrementRateLimit(r, req.Model, req.ApiKey)
	}

	// 残り回数をSSEイベントとして送信
	if rlJSON := h.rateLimitJSON(r, req.Model, req.ApiKey); rlJSON != "" {
		writeSSE(w, "rate_limit", rlJSON)
		if canFlush {
			flusher.Flush()
		}
	}
}

// writeSSE は SSE (Server-Sent Events) 形式でイベントを書き込む
func writeSSE(w http.ResponseWriter, eventType, data string) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, data)
}

// jsonEscape は文字列を JSON 文字列としてエスケープする
func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b[1 : len(b)-1])
}
