package handler

import (
	"context"
	"database/sql"

	pb "github.com/TenshoOHASHI/knowhub/proto/wiki" // 生成されたprotoコード
	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/model"
	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CQRS ハンドラー
type WikiCQRSHandler struct {
	pb.UnimplementedWikiServicesServer
	commandRepo repository.ArticleCommandRepository // Create/Save/Delete
	queryRepo   repository.ArticleQueryRepository   // FindList/FindByList
}

// ハンドラー側で、repoをラップ
// 使う側は、ArticleRepositoryを満たす必要がある
func NewWikiCQRSHandler(commandRepo repository.ArticleCommandRepository, queryRepo repository.ArticleQueryRepository) *WikiCQRSHandler {
	return &WikiCQRSHandler{
		commandRepo: commandRepo,
		queryRepo:   queryRepo,
	}
}

func (h *WikiCQRSHandler) Create(ctx context.Context, req *pb.CreateArticleRequest) (*pb.CreateArticleResponse, error) {
	// ドメインのバリデーションチェック
	// 記事のインスタンスを作成
	article, err := model.NewArticle(req.Title, req.Content)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = h.commandRepo.Create(ctx, article)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create article")
	}

	// 基底クラスで、データを返す
	return &pb.CreateArticleResponse{
		Article: toProductArticle(article),
	}, nil
}

func (h *WikiCQRSHandler) Get(ctx context.Context, req *pb.GetArticleRequest) (*pb.GetArticleResponse, error) {
	article, err := h.queryRepo.FindById(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			// 404
			return nil, status.Error(codes.NotFound, "article not found")
		}
		// 500
		return nil, status.Error(codes.Internal, "failed to get article")
	}

	return &pb.GetArticleResponse{
		Article: toProductArticle(article),
	}, nil
}

func (h *WikiCQRSHandler) List(ctx context.Context, req *pb.ListArticleRequest) (*pb.ListArticleResponse, error) {
	articles, err := h.queryRepo.FindAll(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list article")
	}

	return &pb.ListArticleResponse{
		Article: toProductCQRSArticles(articles),
	}, nil
}

func (h *WikiCQRSHandler) Update(ctx context.Context, req *pb.UpdateArticleRequest) (*pb.UpdateArticleResponse, error) {
	// すでに存在しているか確認
	article, err := h.queryRepo.FindById(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "article not found")
		}
		return nil, status.Error(codes.Internal, "failed to find article")
	}

	// NOTE: optionalを指定しているため、*string(nilになる可能性がある)
	var title, content string // もしnilの場合は、ゼロ値の空文字で初期化される
	if req.Title != nil {
		title = *req.Title
	}

	if req.Content != nil {
		content = *req.Content
	}

	// ここの引数はポインターアドレスを渡す必要があるみたいです？
	article.Update(title, content) // メソッドを呼び出し、既存の値を上書き
	err = h.commandRepo.Save(ctx, article)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update article")
	}

	return &pb.UpdateArticleResponse{
		Article: toProductCQRSArticle(article),
	}, nil
}

func (h *WikiCQRSHandler) Delete(ctx context.Context, req *pb.DeleteArticleRequest) (*emptypb.Empty, error) {
	err := h.commandRepo.Delete(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete article")
	}
	return &emptypb.Empty{}, nil
}

// helper
// modelからprotoに変換
func toProductCQRSArticle(a *model.Article) *pb.Article {
	return &pb.Article{
		Id:        a.ID,
		Title:     a.Title,
		Content:   a.Content,
		CreatedAt: timestamppb.New(a.CreatedAt),
		UpdatedAt: timestamppb.New(a.UpdatedAt),
	}
}

// modelからprotoに変換
func toProductCQRSArticles(articles []*model.Article) []*pb.Article {
	var result []*pb.Article
	for _, a := range articles {
		result = append(result, toProductArticle(a))
	}

	return result
}
