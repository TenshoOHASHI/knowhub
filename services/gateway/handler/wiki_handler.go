package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

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
	client pb.WikiServicesClient // gRPCクライアント
}

func NewWikiHandler(conn *grpc.ClientConn) *WikiHandler {
	return &WikiHandler{
		client: pb.NewWikiServicesClient(conn),
	}
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
}

// UpdateArticle 記事更新
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
}

// DeleteArticle 記事削除
// @Summary      記事削除
// @Description  指定したIDの記事を削除する
// @Tags         wiki
// @Param        id   path      string  true  "記事ID"
// @Success      204  "No Content"
// @Failure      500  {string}  string  "internal server error"
// @Router       /api/articles/{id} [delete]
func (h *WikiHandler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

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
