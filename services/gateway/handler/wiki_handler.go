package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/TenshoOHASHI/knowhub/pkg/notifier"
	aiPb "github.com/TenshoOHASHI/knowhub/proto/ai"
	pb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"github.com/TenshoOHASHI/knowhub/services/gateway/swagger"
	"google.golang.org/grpc"
)

// swagger types (used by swag annotations)
var (
	_ swagger.CreateArticleRequest
	_ swagger.UpdateArticleRequest
	_ swagger.CreateCategoryRequest
)

type WikiHandler struct {
	client   pb.WikiServicesClient // gRPCクライアント
	aiClient aiPb.AIServiceClient  // グラフキャッシュ無効化用
	notifier *notifier.SlackNotifier
}

func NewWikiHandler(conn *grpc.ClientConn) *WikiHandler {
	return &WikiHandler{
		client: pb.NewWikiServicesClient(conn),
	}
}

// Client returns the underlying gRPC client for use by other handlers
func (h *WikiHandler) Client() pb.WikiServicesClient {
	return h.client
}

// SetNotifier は Slack Notifier を設定する
func (h *WikiHandler) SetNotifier(n *notifier.SlackNotifier) {
	h.notifier = n
}

// SetAIClient は AI Service クライアントを設定する（記事CRUD時のグラフキャッシュ無効化用）
func (h *WikiHandler) SetAIClient(aiClient aiPb.AIServiceClient) {
	h.aiClient = aiClient
}

// invalidateGraphCache は AI Service のグラフキャッシュを非同期で無効化する
func (h *WikiHandler) invalidateGraphCache() {
	if h.aiClient == nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// GetKnowledgeGraph を refresh=true 付きで呼び出すことでキャッシュを無効化
		// → 次回アクセス時に自動再構築される
		_, _ = h.aiClient.InvalidateGraphCache(ctx, &aiPb.InvalidateGraphCacheRequest{})
		slog.Info("graph cache invalidation requested")
	}()
}

// ListArticles 記事一覧取得
// @Summary      記事一覧取得
// @Description  全ての記事を新しい順で返す
// @Tags         wiki
// @Produce      json
// @Success      200  {array}   map[string]interface{}
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/articles [get]
func (h *WikiHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	// サーバ内部のぷろせじゃーでサービスを呼んで、値を返しているってこと？
	resp, err := h.client.List(r.Context(), &pb.ListArticleRequest{})
	if err != nil {
		slog.Error("failed to list articles", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	// Go構造体　-> JSON文字列 -> ResW -> ネットワークに書き込む
	json.NewEncoder(w).Encode(resp)

}

// GetArticle 記事詳細取得
// @Summary      記事詳細取得
// @Description  指定したIDの記事を返す
// @Tags         wiki
// @Produce      json
// @Param        id   path      string  true  "記事ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/articles/{id} [get]
func (h *WikiHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	resp, err := h.client.Get(r.Context(), &pb.GetArticleRequest{
		Id: id,
	})
	if err != nil {
		slog.Error("failed to get article", "error", err, "id", id)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// 非公開記事は認証済みユーザーのみアクセス可能
	if resp.Article.Visibility == "locked" {
		userID := r.Context().Value("userID")
		if userID == nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateArticle 記事作成
// @Summary      記事作成
// @Description  新しい記事を作成する
// @Tags         wiki
// @Accept       json
// @Produce      json
// @Param        request  body  swagger.CreateArticleRequest  true  "記事データ"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {string}  string  "invalid request body"
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/articles [post]
func (h *WikiHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	// リクエスト用の構造体を用意
	// Wiki Service:  「titleが1文字以上か」→ model.NewArticle() で弾く
	var req struct {
		Title      string `json:"title"`
		Content    string `json:"content"`
		CategoryID string `json:"category_id"`
		Visibility string `json:"visibility"`
	}

	// リクエストbodyをjsonからGo構造体に変換
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// gRPC
	resp, err := h.client.Create(r.Context(), &pb.CreateArticleRequest{
		Title:      req.Title,
		Content:    req.Content,
		CategoryId: req.CategoryID,
		Visibility: req.Visibility,
	})

	if err != nil {
		slog.Error("failed to create article", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)

	// 記事が追加されたのでグラフキャッシュを無効化
	h.invalidateGraphCache()

	// Slack通知
	if h.notifier != nil && resp.Article != nil {
		h.notifier.NotifyAsync(func() error {
			return h.notifier.NotifyArticleCreated(resp.Article.Title, resp.Article.Id)
		})
	}
}

// @Summary      記事更新
// @Description  指定したIDの記事を更新する（部分更新対応）
// @Tags         wiki
// @Accept       json
// @Produce      json
// @Param        id       path      string  true  "記事ID"
// @Param        request  body      swagger.UpdateArticleRequest  true  "更新データ"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {string}  string  "invalid request body"
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/articles/{id} [put]
func (h *WikiHandler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		// ポインター型にすることで、無いと空を区別する
		// もし、値がなければ、nilになる、その場合、そのフィールドは更新しない
		Title      *string `json:"title,omitempty"` // 空値ならjsonに出力しない
		Content    *string `json:"content,omitempty"`
		Visibility *string `json:"visibility,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.Update(r.Context(), &pb.UpdateArticleRequest{
		Id:         id,
		Title:      req.Title,
		Content:    req.Content, // ポインタアドレスで指定する必要がある、nilになる可能性があるため
		Visibility: req.Visibility,
	})

	if err != nil {
		slog.Error("failed to update article", "error", err, "id", id)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)

	// 記事が更新されたのでグラフキャッシュを無効化
	h.invalidateGraphCache()

	// Slack通知
	if h.notifier != nil && resp.Article != nil {
		h.notifier.NotifyAsync(func() error {
			return h.notifier.NotifyArticleUpdated(resp.Article.Title, resp.Article.Id)
		})
	}
}

// @Summary      記事削除
// @Description  指定したIDの記事を削除する
// @Tags         wiki
// @Param        id   path      string  true  "記事ID"
// @Success      204  "No Content"
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/articles/{id} [delete]
func (h *WikiHandler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Slack通知のために削除前にタイトルを取得
	var articleTitle string
	if h.notifier != nil {
		getResp, err := h.client.Get(r.Context(), &pb.GetArticleRequest{Id: id})
		if err == nil && getResp.Article != nil {
			articleTitle = getResp.Article.Title
		}
	}

	_, err := h.client.Delete(r.Context(), &pb.DeleteArticleRequest{
		Id: id,
	})
	if err != nil {
		slog.Error("failed to delete article", "error", err, "id", id)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

	h.invalidateGraphCache()

	// Slack通知
	if h.notifier != nil && articleTitle != "" {
		h.notifier.NotifyAsync(func() error {
			return h.notifier.NotifyArticleDeleted(articleTitle, id)
		})
	}
}

// ===== Category =====

// ListCategories カテゴリ一覧取得
// @Summary      カテゴリ一覧取得
// @Description  全てのカテゴリを階層構造で返す
// @Tags         categories
// @Produce      json
// @Success      200  {array}   map[string]interface{}
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/categories [get]
func (h *WikiHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.ListCategories(r.Context(), &pb.ListCategoriesRequest{})
	if err != nil {
		slog.Error("failed to list categories", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateCategory カテゴリ作成
// @Summary      カテゴリ作成
// @Description  新しいカテゴリを作成する（parent_id で階層化可能）
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        request  body  swagger.CreateCategoryRequest  true  "カテゴリデータ"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {string}  string  "invalid request body"
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/categories [post]
func (h *WikiHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		ParentID string `json:"parent_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.CreateCategory(r.Context(), &pb.CreateCategoryRequest{
		Name:     req.Name,
		ParentId: req.ParentID,
	})
	if err != nil {
		slog.Error("failed to create category", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// DeleteCategory カテゴリ削除
// @Summary      カテゴリ削除
// @Description  指定したIDのカテゴリを削除する
// @Tags         categories
// @Param        id   path      string  true  "カテゴリID"
// @Success      204  "No Content"
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/categories/{id} [delete]
func (h *WikiHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	_, err := h.client.DeleteCategory(r.Context(), &pb.DeleteCategoryRequest{
		Id: id,
	})
	if err != nil {
		slog.Error("failed to delete category", "error", err, "id", id)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ===== Like / Save =====

func validateFingerprint(fingerprint string) bool {
	fingerprint = strings.TrimSpace(fingerprint)
	if len(fingerprint) != 64 {
		return false
	}
	for _, r := range fingerprint {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

// ToggleLike いいね追加/解除
func (h *WikiHandler) ToggleLike(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Fingerprint string `json:"fingerprint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if !validateFingerprint(req.Fingerprint) {
		http.Error(w, "invalid fingerprint", http.StatusBadRequest)
		return
	}
	resp, err := h.client.ToggleLike(r.Context(), &pb.ToggleLikeRequest{
		ArticleId:   id,
		Fingerprint: req.Fingerprint,
	})
	if err != nil {
		slog.Error("failed to toggle like", "error", err, "id", id)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetLikeCount いいね数取得
func (h *WikiHandler) GetLikeCount(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fingerprint := r.URL.Query().Get("fingerprint")
	if !validateFingerprint(fingerprint) {
		http.Error(w, "invalid fingerprint", http.StatusBadRequest)
		return
	}
	resp, err := h.client.GetLikeCount(r.Context(), &pb.GetLikeCountRequest{
		ArticleId:   id,
		Fingerprint: fingerprint,
	})
	if err != nil {
		slog.Error("failed to get like count", "error", err, "id", id)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetLikeCounts 複数記事のいいね数一括取得
func (h *WikiHandler) GetLikeCounts(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ArticleIDs []string `json:"article_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	resp, err := h.client.GetLikeCounts(r.Context(), &pb.GetLikeCountsRequest{
		ArticleIds: req.ArticleIDs,
	})
	if err != nil {
		slog.Error("failed to get like counts", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// SaveArticle 記事保存
func (h *WikiHandler) SaveArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Fingerprint string `json:"fingerprint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if !validateFingerprint(req.Fingerprint) {
		http.Error(w, "invalid fingerprint", http.StatusBadRequest)
		return
	}
	resp, err := h.client.SaveArticle(r.Context(), &pb.SaveArticleRequest{
		ArticleId:   id,
		Fingerprint: req.Fingerprint,
	})
	if err != nil {
		slog.Error("failed to save article", "error", err, "id", id)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UnsaveArticle 保存解除
func (h *WikiHandler) UnsaveArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Fingerprint string `json:"fingerprint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if !validateFingerprint(req.Fingerprint) {
		http.Error(w, "invalid fingerprint", http.StatusBadRequest)
		return
	}
	resp, err := h.client.UnsaveArticle(r.Context(), &pb.UnsaveArticleRequest{
		ArticleId:   id,
		Fingerprint: req.Fingerprint,
	})
	if err != nil {
		slog.Error("failed to unsave article", "error", err, "id", id)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListSavedArticles 保存済み記事一覧
func (h *WikiHandler) ListSavedArticles(w http.ResponseWriter, r *http.Request) {
	fingerprint := r.URL.Query().Get("fingerprint")
	if !validateFingerprint(fingerprint) {
		http.Error(w, "invalid fingerprint", http.StatusBadRequest)
		return
	}
	resp, err := h.client.ListSavedArticles(r.Context(), &pb.ListSavedArticlesRequest{
		Fingerprint: fingerprint,
	})
	if err != nil {
		slog.Error("failed to list saved articles", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// IsArticleSaved 保存済みか確認
func (h *WikiHandler) IsArticleSaved(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fingerprint := r.URL.Query().Get("fingerprint")
	if !validateFingerprint(fingerprint) {
		http.Error(w, "invalid fingerprint", http.StatusBadRequest)
		return
	}
	resp, err := h.client.IsArticleSaved(r.Context(), &pb.IsArticleSavedRequest{
		ArticleId:   id,
		Fingerprint: fingerprint,
	})
	if err != nil {
		slog.Error("failed to check saved status", "error", err, "id", id)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
