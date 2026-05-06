package handler

import (
	"context"
	"encoding/json"
	"fmt"
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

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AskWithAgentStream — POST /api/ai/agent/stream (SSE)
//
// 処理フロー (直列):
//  1. SSE ヘッダーを設定して HTTP 接続を維持
//  2. gRPC サーバーストリーミングを開始 (AI Service に接続)
//  3. Recv() でイベントを1つずつ受信 → writeSSE() で SSE 形式に変換 → Flush() で即座に送信
//  4. Recv() が io.EOF を返したらストリーム終了
//
// Flush() が重要: これがないと ResponseWriter のバッファに溜まり、
// ブラウザに一気に届いてリアルタイム性が失われる。
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

	// SSE ヘッダー設定
	// text/event-stream: ブラウザに「これはSSEストリームだ」と伝える
	// no-cache: プロキシ/CDNがレスポンスをキャッシュしないようにする
	// keep-alive: HTTP接続を維持してサーバーから複数回データを送れるようにする
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // nginx バッファリング無効化

	ctx, cancel := context.WithTimeout(r.Context(), 300*time.Second)
	defer cancel()

	// gRPC サーバーストリーミング開始
	// AI Service との間に持続的な接続が確立される
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

	// 直列処理ループ: AI Service からのイベントを1つずつ受信 → SSE変換 → 送信
	for {
		// Recv() は次のイベントが届くまでブロックする (HTTP/2 フレームとして届く)
		// io.EOF = ストリーム終了 (AI Service が全イベント送信完了)
		event, err := stream.Recv()
		if err != nil {
			break
		}

		data, err := json.Marshal(event)
		if err != nil {
			continue
		}

		// SSE 形式で書き出し:
		//   "event: step\ndata: {...}\n\n"
		// \n\n (空行) がメッセージの区切り
		writeSSE(w, event.EventType, string(data))

		// Flush() で ResponseWriter のバッファを即座にブラウザに送信
		// これがないとバッファに溜まり、リアルタイム表示ができない
		if canFlush {
			flusher.Flush()
		}
	}
}

// writeSSE は SSE (Server-Sent Events) 形式でイベントを書き込む
//
// SSE メッセージフォーマット:
//
//	event: <イベント種別>\n
//	data: <JSONペイロード>\n
//	\n    ← 空行がメッセージの終了を示す
//
// 例: writeSSE(w, "step", `{"phase":"llm_thinking"}`)
//
//	→ "event: step\ndata: {"phase":"llm_thinking"}\n\n"
func writeSSE(w http.ResponseWriter, eventType, data string) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, data)
}

// jsonEscape は文字列を JSON 文字列としてエスケープする
func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b[1 : len(b)-1])
}
