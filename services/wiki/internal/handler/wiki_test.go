package handler_test

import (
	"context"
	"database/sql"
	"testing"

	pb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/handler"
	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/model"
)

type mockRepository struct {
	// id=valueでテーブルを真似る
	// {"id1": {data}, "id2": "data" }
	articles map[string]*model.Article
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		articles: make(map[string]*model.Article),
	}
}
func (m *mockRepository) Create(ctx context.Context, article *model.Article) error {
	m.articles[article.ID] = article
	return nil
}

func (m *mockRepository) FindById(ctx context.Context, id string) (*model.Article, error) {
	article, exists := m.articles[id]
	if !exists {
		return nil, sql.ErrNoRows
	}
	return article, nil
}

func (m *mockRepository) FindAll(ctx context.Context) ([]*model.Article, error) {
	var articles []*model.Article
	for _, a := range m.articles {
		articles = append(articles, a)
	}
	return articles, nil
}

func (m *mockRepository) Save(ctx context.Context, article *model.Article) error {
	m.articles[article.ID] = article
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	delete(m.articles, id)
	return nil
}

func TestCreate_Success(t *testing.T) {
	repo := newMockRepository()
	h := handler.NewWikiHandler(repo)

	resp, err := h.Create(context.Background(), &pb.CreateArticleRequest{
		Title:   "Go入門",
		Content: "テスト",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Article.Title != "Go入門" {
		t.Errorf("expected title Go入門, got %s", resp.Article.Title)
	}

	// モックの中にも保存されているか
	if len(repo.articles) != 1 {
		t.Errorf("expected 1 article, got %d", len(repo.articles))
	}
}

func TestGet_NotFound(t *testing.T) {
	repo := newMockRepository()
	h := handler.NewWikiHandler(repo)

	// からのIDの場合NotFoundになるほず
	reps, err := h.Get(context.Background(), &pb.GetArticleRequest{
		Id: "存在しない",
	})

	// 空であるべき
	if reps != nil {
		t.Error("expected nil response")
	}

	// gRPCのステートコードを取得
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected gRPC error")
	}
	// ステートコードが、NotFoundになるべき
	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound, got %s", st.Code())
	}

}
