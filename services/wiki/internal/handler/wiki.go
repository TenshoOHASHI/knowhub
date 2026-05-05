package handler

import (
	"context"
	"database/sql"
	"log/slog"

	pb "github.com/TenshoOHASHI/knowhub/proto/wiki" // 生成されたprotoコード
	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/model"
	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// WikiHandler はgRPCのハンドラー
type WikiHandler struct {
	pb.UnimplementedWikiServicesServer                              // protoから生成された基底構造体
	repo                               repository.ArticleRepository // interface（テスト時にモック差し替え可）
}

// ハンドラー側で、repoをラップ
// 使う側は、ArticleRepositoryを満たす必要がある
func NewWikiHandler(repo repository.ArticleRepository) *WikiHandler {
	return &WikiHandler{
		repo: repo,
	}
}

func (h *WikiHandler) Create(ctx context.Context, req *pb.CreateArticleRequest) (*pb.CreateArticleResponse, error) {
	// ドメインのバリデーションチェック
	// 記事のインスタンスを作成
	article, err := model.NewArticle(req.Title, req.Content, req.CategoryId, req.Visibility)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = h.repo.Create(ctx, article)
	if err != nil {
		slog.Error("failed to create article", "error", err, "title", req.Title)
		return nil, status.Error(codes.Internal, "failed to create article")
	}

	// 基底クラスで、データを返す
	return &pb.CreateArticleResponse{
		Article: toProductArticle(article),
	}, nil
}

func (h *WikiHandler) Get(ctx context.Context, req *pb.GetArticleRequest) (*pb.GetArticleResponse, error) {
	article, err := h.repo.FindById(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			// 404
			return nil, status.Error(codes.NotFound, "article not found")
		}
		// 500
		slog.Error("failed to get article", "error", err, "id", req.Id)
		return nil, status.Error(codes.Internal, "failed to get article")
	}

	return &pb.GetArticleResponse{
		Article: toProductArticle(article),
	}, nil
}

func (h *WikiHandler) List(ctx context.Context, req *pb.ListArticleRequest) (*pb.ListArticleResponse, error) {
	articles, err := h.repo.FindAll(ctx)
	if err != nil {
		slog.Error("failed to list articles", "error", err)
		return nil, status.Error(codes.Internal, "failed to list article")
	}

	return &pb.ListArticleResponse{
		Article: toProductArticles(articles),
	}, nil
}

func (h *WikiHandler) Update(ctx context.Context, req *pb.UpdateArticleRequest) (*pb.UpdateArticleResponse, error) {
	// すでに存在しているか確認
	article, err := h.repo.FindById(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "article not found")
		}
		slog.Error("failed to find article for update", "error", err, "id", req.Id)
		return nil, status.Error(codes.Internal, "failed to find article")
	}

	// NOTE: optionalを指定しているため、*string(nilになる可能性がある)
	var title, content, visibility string // もしnilの場合は、ゼロ値の空文字で初期化される
	if req.Title != nil {
		title = *req.Title
	}

	if req.Content != nil {
		content = *req.Content
	}

	if req.Visibility != nil {
		visibility = *req.Visibility
	}

	// ここの引数はポインターアドレスを渡す必要があるみたいです？
	article.Update(title, content, visibility) // メソッドを呼び出し、既存の値を上書き
	err = h.repo.Save(ctx, article)
	if err != nil {
		slog.Error("failed to save article", "error", err, "id", req.Id)
		return nil, status.Error(codes.Internal, "failed to update article")
	}

	return &pb.UpdateArticleResponse{
		Article: toProductArticle(article),
	}, nil
}

func (h *WikiHandler) Delete(ctx context.Context, req *pb.DeleteArticleRequest) (*emptypb.Empty, error) {
	err := h.repo.Delete(ctx, req.Id)
	if err != nil {
		slog.Error("failed to delete article", "error", err, "id", req.Id)
		return nil, status.Error(codes.Internal, "failed to delete article")
	}
	return &emptypb.Empty{}, nil
}

// helper
// modelからprotoに変換
func toProductArticle(a *model.Article) *pb.Article {
	return &pb.Article{
		Id:         a.ID,
		Title:      a.Title,
		Content:    a.Content,
		CategoryId: a.CategoryID,
		Visibility: a.Visibility,
		CreatedAt:  timestamppb.New(a.CreatedAt),
		UpdatedAt:  timestamppb.New(a.UpdatedAt),
	}
}

// modelからprotoに変換
func toProductArticles(articles []*model.Article) []*pb.Article {
	var result []*pb.Article
	for _, a := range articles {
		result = append(result, toProductArticle(a))
	}

	return result
}
