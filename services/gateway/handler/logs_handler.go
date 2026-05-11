package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// DockerOperator は DockerClient が実装するインターフェース
type DockerOperator interface {
	StreamLogs(ctx context.Context, services []string) (io.ReadCloser, error)
	ListContainers(ctx context.Context) ([]ContainerJSON, error)
	ExecCommand(ctx context.Context, service, action string) (string, error)
}

// ContainerJSON はコンテナ情報のJSON表現
type ContainerJSON struct {
	Name    string `json:"name"`
	Service string `json:"service"`
	State   string `json:"state"`
	Status  string `json:"status"`
}

// LogsHandler はログ監視ダッシュボードのハンドラー
type LogsHandler struct {
	docker DockerOperator
}

// NewLogsHandler は LogsHandler を作成する
func NewLogsHandler(docker DockerOperator) *LogsHandler {
	return &LogsHandler{docker: docker}
}

// ログレベルの優先度
var logLevelPriority = map[string]int{
	"DEBUG": 0,
	"INFO":  1,
	"WARN":  2,
	"ERROR": 3,
}

// shouldIncludeLog はログ行が指定レベル以上かどうかを判定する
func shouldIncludeLog(line string, minLevel string) bool {
	if minLevel == "" || minLevel == "all" {
		return true
	}

	minPriority, ok := logLevelPriority[strings.ToUpper(minLevel)]
	if !ok {
		return true
	}

	// slog JSON フォーマット: {"level":"INFO", ...}
	// 高速判定: 行内の "level":"XXX" パターンを探す
	for level, priority := range logLevelPriority {
		if strings.Contains(line, `"level":"`+level+`"`) {
			return priority >= minPriority
		}
	}

	// JSON パースできないログ行（docker compose のプレフィックス行等）は常に含める
	return true
}

// StreamLogsHandler は SSE でログをストリーミングする
// GET /api/logs/stream?services=ai,gateway&level=error
func (h *LogsHandler) StreamLogsHandler(w http.ResponseWriter, r *http.Request) {
	// クエリパラメータ解析
	servicesParam := r.URL.Query().Get("services")
	levelParam := r.URL.Query().Get("level")

	var services []string
	if servicesParam != "" {
		services = strings.Split(servicesParam, ",")
	}

	slog.Info("logs: stream started",
		"services", services,
		"level", levelParam,
		"remote", r.RemoteAddr)

	// SSE ヘッダー設定
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Nginx バッファリング無効化

	flusher, canFlush := w.(http.Flusher)

	// リクエストの context を使用（クライアント切断時に自動キャンセル）
	ctx := r.Context()

	reader, err := h.docker.StreamLogs(ctx, services)
	if err != nil {
		slog.Error("logs: failed to start streaming", "error", err)
		writeSSELog(w, "error", fmt.Sprintf(`{"message":"failed to start log streaming: %s"}`, err.Error()))
		if canFlush {
			flusher.Flush()
		}
		return
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	// デフォルトのバッファサイズを拡張（長いログ行に対応）
	scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)

	for scanner.Scan() {
		// context がキャンセルされたら終了
		if ctx.Err() != nil {
			break
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		// サーバーサイドログレベルフィルタリング
		if !shouldIncludeLog(line, levelParam) {
			continue
		}

		// SSE イベントとして送信
		writeSSELog(w, "log", line)
		if canFlush {
			flusher.Flush()
		}
	}

	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		slog.Warn("logs: scanner error", "error", err)
	}

	slog.Info("logs: stream ended", "services", services)
}

// ListContainersHandler はコンテナ一覧を返す
// GET /api/logs/containers
func (h *LogsHandler) ListContainersHandler(w http.ResponseWriter, r *http.Request) {
	containers, err := h.docker.ListContainers(r.Context())
	if err != nil {
		slog.Error("logs: failed to list containers", "error", err)
		http.Error(w, "failed to list containers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containers)
}

// ExecActionHandler はDocker操作を実行する
// POST /api/logs/action
// Body: {"service": "nginx", "action": "reload"}
func (h *LogsHandler) ExecActionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Service string `json:"service"`
		Action  string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Service == "" || req.Action == "" {
		http.Error(w, "service and action are required", http.StatusBadRequest)
		return
	}

	slog.Info("logs: executing action",
		"service", req.Service,
		"action", req.Action,
		"remote", r.RemoteAddr)

	output, err := h.docker.ExecCommand(r.Context(), req.Service, req.Action)
	if err != nil {
		slog.Error("logs: action failed",
			"service", req.Service,
			"action", req.Action,
			"error", err,
			"output", output)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":  err.Error(),
			"output": output,
		})
		return
	}

	slog.Info("logs: action completed",
		"service", req.Service,
		"action", req.Action,
		"output", output)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"output": output,
	})
}

// writeSSELog は SSE 形式でログイベントを書き込む
func writeSSELog(w http.ResponseWriter, eventType, data string) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, data)
}
