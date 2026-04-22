package handler

import (
	"encoding/json"
	"net/http"

	pb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"google.golang.org/grpc"
)

type WikiHandler struct {
	client pb.WikiServicesClient // gRPCクライアント
}

func NewWikiHandler(conn *grpc.ClientConn) *WikiHandler {
	return &WikiHandler{
		client: pb.NewWikiServicesClient(conn),
	}
}

// GET /api/articles
func (h *WikiHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	// サーバ内部のぷろせじゃーでサービスを呼んで、値を返しているってこと？
	resp, err := h.client.List(r.Context(), &pb.ListArticleRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	// Go構造体　-> JSON文字列 -> ResW -> ネットワークに書き込む
	json.NewEncoder(w).Encode(resp)

}

// Get /api/articles/{id}
func (h *WikiHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	resp, err := h.client.Get(r.Context(), &pb.GetArticleRequest{
		Id: id,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// mapにキーと値を格納（json形式で通信すると明記）
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *WikiHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	// リクエスト用の構造体を用意
	// Wiki Service:  「titleが1文字以上か」→ model.NewArticle() で弾く
	var req struct {
		Title      string `json:"title"`
		Content    string `json:"content"`
		CategoryID string `json:"category_id"`
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
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // ヘッダーにステータスコードを追加、201
	json.NewEncoder(w).Encode(resp)
}

func (h *WikiHandler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		// ポインター型にすることで、無いと空を区別する
		// もし、値がなければ、nilになる、その場合、そのフィールドは更新しない
		Title   *string `json:"title,omitempty"` // 空値ならjsonに出力しない
		Content *string `json:"content,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.client.Update(r.Context(), &pb.UpdateArticleRequest{
		Id:      id,
		Title:   req.Title,
		Content: req.Content, // ポインタアドレスで指定する必要がある、nilになる可能性があるため
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *WikiHandler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	_, err := h.client.Delete(r.Context(), &pb.DeleteArticleRequest{
		Id: id,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent) // 204 =　Bodyなし
}

// ===== Category =====

// GET /api/categories
func (h *WikiHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.ListCategories(r.Context(), &pb.ListCategoriesRequest{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// POST /api/categories
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// DELETE /api/categories/{id}
func (h *WikiHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	_, err := h.client.DeleteCategory(r.Context(), &pb.DeleteCategoryRequest{
		Id: id,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
