package handler

import (
	"context"
	"database/sql"
	"log/slog"

	pb "github.com/TenshoOHASHI/knowhub/proto/wiki" // 生成されたprotoコード
	// 生成されたprotoコード
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
	commandRepo      repository.ArticleCommandRepository // Create/Save/Delete
	queryRepo        repository.ArticleQueryRepository   // FindList/FindByList
	categoryRepo     repository.CategoryRepository       // Category CRUD
	likeRepo         repository.LikeRepository
	savedArticleRepo repository.SavedArticleRepository
	analyticsRepo    repository.AnalyticsRepository
}

// ハンドラー側で、repoをラップ
// 使う側は、ArticleRepositoryを満たす必要がある
func NewWikiCQRSHandler(commandRepo repository.ArticleCommandRepository, queryRepo repository.ArticleQueryRepository, categoryRepo repository.CategoryRepository, likeRepo repository.LikeRepository, savedArticleRepo repository.SavedArticleRepository, analyticsRepo repository.AnalyticsRepository) *WikiCQRSHandler {
	return &WikiCQRSHandler{
		commandRepo:      commandRepo,
		queryRepo:        queryRepo,
		categoryRepo:     categoryRepo,
		likeRepo:         likeRepo,
		savedArticleRepo: savedArticleRepo,
		analyticsRepo:    analyticsRepo,
	}
}

func (h *WikiCQRSHandler) Create(ctx context.Context, req *pb.CreateArticleRequest) (*pb.CreateArticleResponse, error) {
	// ドメインのバリデーションチェック
	// 記事のインスタンスを作成
	article, err := model.NewArticle(req.Title, req.Content, req.CategoryId, req.Visibility)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = h.commandRepo.Create(ctx, article)
	if err != nil {
		slog.Error("failed to create article", "error", err, "title", req.Title)
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
		slog.Error("failed to get article", "error", err, "id", req.Id)
		return nil, status.Error(codes.Internal, "failed to get article")
	}

	return &pb.GetArticleResponse{
		Article: toProductArticle(article),
	}, nil
}

func (h *WikiCQRSHandler) List(ctx context.Context, req *pb.ListArticleRequest) (*pb.ListArticleResponse, error) {
	articles, err := h.queryRepo.FindAll(ctx)
	if err != nil {
		slog.Error("failed to list articles", "error", err)
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
	err = h.commandRepo.Save(ctx, article)
	if err != nil {
		slog.Error("failed to save article", "error", err, "id", req.Id)
		return nil, status.Error(codes.Internal, "failed to update article")
	}

	return &pb.UpdateArticleResponse{
		Article: toProductCQRSArticle(article),
	}, nil
}

func (h *WikiCQRSHandler) Delete(ctx context.Context, req *pb.DeleteArticleRequest) (*emptypb.Empty, error) {
	err := h.commandRepo.Delete(ctx, req.Id)
	if err != nil {
		slog.Error("failed to delete article", "error", err, "id", req.Id)
		return nil, status.Error(codes.Internal, "failed to delete article")
	}
	return &emptypb.Empty{}, nil
}

// helper
// modelからprotoに変換
func toProductCQRSArticle(a *model.Article) *pb.Article {
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
func toProductCQRSArticles(articles []*model.Article) []*pb.Article {
	var result []*pb.Article
	for _, a := range articles {
		result = append(result, toProductArticle(a))
	}

	return result
}

// ===== Category RPC =====

func (h *WikiCQRSHandler) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	categories, err := h.categoryRepo.FindAll(ctx)
	if err != nil {
		slog.Error("failed to list categories", "error", err)
		return nil, status.Error(codes.Internal, "failed to list categories")
	}

	var pbCategories []*pb.Category
	for _, c := range categories {
		pbCategories = append(pbCategories, &pb.Category{
			Id:       c.ID,
			Name:     c.Name,
			ParentId: c.ParentID,
		})
	}

	return &pb.ListCategoriesResponse{
		Categories: pbCategories,
	}, nil
}

func (h *WikiCQRSHandler) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	category, err := model.NewCategory(req.Name, req.ParentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = h.categoryRepo.Create(ctx, category)
	if err != nil {
		slog.Error("failed to create category", "error", err, "name", req.Name)
		return nil, status.Error(codes.Internal, "failed to create category")
	}

	return &pb.CreateCategoryResponse{
		Category: &pb.Category{
			Id:       category.ID,
			Name:     category.Name,
			ParentId: category.ParentID,
		},
	}, nil
}

func (h *WikiCQRSHandler) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*emptypb.Empty, error) {
	err := h.categoryRepo.Delete(ctx, req.Id)
	if err != nil {
		slog.Error("failed to delete category", "error", err, "id", req.Id)
		return nil, status.Error(codes.Internal, "failed to delete category")
	}
	return &emptypb.Empty{}, nil
}

// ===== Like / Save RPC =====

func (h *WikiCQRSHandler) ToggleLike(ctx context.Context, req *pb.ToggleLikeRequest) (*pb.ToggleLikeResponse, error) {
	if req.ArticleId == "" || req.Fingerprint == "" {
		return nil, status.Error(codes.InvalidArgument, "article_id and fingerprint are required")
	}
	count, liked, err := h.likeRepo.ToggleLike(ctx, req.ArticleId, req.Fingerprint)
	if err != nil {
		slog.Error("failed to toggle like", "error", err, "article_id", req.ArticleId)
		return nil, status.Error(codes.Internal, "failed to toggle like")
	}
	return &pb.ToggleLikeResponse{Count: count, Liked: liked}, nil
}

func (h *WikiCQRSHandler) GetLikeCount(ctx context.Context, req *pb.GetLikeCountRequest) (*pb.GetLikeCountResponse, error) {
	if req.ArticleId == "" {
		return nil, status.Error(codes.InvalidArgument, "article_id is required")
	}
	count, liked, err := h.likeRepo.GetLikeCount(ctx, req.ArticleId, req.Fingerprint)
	if err != nil {
		slog.Error("failed to get like count", "error", err, "article_id", req.ArticleId)
		return nil, status.Error(codes.Internal, "failed to get like count")
	}
	return &pb.GetLikeCountResponse{Count: count, Liked: liked}, nil
}

func (h *WikiCQRSHandler) GetLikeCounts(ctx context.Context, req *pb.GetLikeCountsRequest) (*pb.GetLikeCountsResponse, error) {
	counts, err := h.likeRepo.GetLikeCounts(ctx, req.ArticleIds)
	if err != nil {
		slog.Error("failed to get like counts", "error", err)
		return nil, status.Error(codes.Internal, "failed to get like counts")
	}
	var pbCounts []*pb.ArticleLikeCount
	for _, id := range req.ArticleIds {
		pbCounts = append(pbCounts, &pb.ArticleLikeCount{
			ArticleId: id,
			Count:     counts[id],
		})
	}
	return &pb.GetLikeCountsResponse{Counts: pbCounts}, nil
}

func (h *WikiCQRSHandler) SaveArticle(ctx context.Context, req *pb.SaveArticleRequest) (*pb.SaveArticleResponse, error) {
	if req.ArticleId == "" || req.Fingerprint == "" {
		return nil, status.Error(codes.InvalidArgument, "article_id and fingerprint are required")
	}
	err := h.savedArticleRepo.Save(ctx, req.ArticleId, req.Fingerprint)
	if err != nil {
		slog.Error("failed to save article", "error", err, "article_id", req.ArticleId)
		return nil, status.Error(codes.Internal, "failed to save article")
	}
	return &pb.SaveArticleResponse{Saved: true}, nil
}

func (h *WikiCQRSHandler) UnsaveArticle(ctx context.Context, req *pb.UnsaveArticleRequest) (*pb.UnsaveArticleResponse, error) {
	if req.ArticleId == "" || req.Fingerprint == "" {
		return nil, status.Error(codes.InvalidArgument, "article_id and fingerprint are required")
	}
	err := h.savedArticleRepo.Unsave(ctx, req.ArticleId, req.Fingerprint)
	if err != nil {
		slog.Error("failed to unsave article", "error", err, "article_id", req.ArticleId)
		return nil, status.Error(codes.Internal, "failed to unsave article")
	}
	return &pb.UnsaveArticleResponse{Unsaved: true}, nil
}

func (h *WikiCQRSHandler) ListSavedArticles(ctx context.Context, req *pb.ListSavedArticlesRequest) (*pb.ListSavedArticlesResponse, error) {
	if req.Fingerprint == "" {
		return nil, status.Error(codes.InvalidArgument, "fingerprint is required")
	}
	articles, err := h.savedArticleRepo.ListByFingerprint(ctx, req.Fingerprint)
	if err != nil {
		slog.Error("failed to list saved articles", "error", err)
		return nil, status.Error(codes.Internal, "failed to list saved articles")
	}
	return &pb.ListSavedArticlesResponse{
		Articles: toProductCQRSArticles(articles),
	}, nil
}

func (h *WikiCQRSHandler) IsArticleSaved(ctx context.Context, req *pb.IsArticleSavedRequest) (*pb.IsArticleSavedResponse, error) {
	if req.ArticleId == "" || req.Fingerprint == "" {
		return nil, status.Error(codes.InvalidArgument, "article_id and fingerprint are required")
	}
	saved, err := h.savedArticleRepo.IsSaved(ctx, req.ArticleId, req.Fingerprint)
	if err != nil {
		slog.Error("failed to check if article is saved", "error", err, "article_id", req.ArticleId)
		return nil, status.Error(codes.Internal, "failed to check saved status")
	}
	return &pb.IsArticleSavedResponse{Saved: saved}, nil
}

// ===== Analytics RPC =====

func (h *WikiCQRSHandler) RecordPageView(ctx context.Context, req *pb.RecordPageViewRequest) (*pb.RecordPageViewResponse, error) {
	if req.Path == "" {
		return nil, status.Error(codes.InvalidArgument, "path is required")
	}
	err := h.analyticsRepo.RecordPageView(ctx, req.Path, req.IpHash, req.UserAgent, req.Referrer)
	if err != nil {
		slog.Error("failed to record page view", "error", err, "path", req.Path)
		return nil, status.Error(codes.Internal, "failed to record page view")
	}
	return &pb.RecordPageViewResponse{}, nil
}

func (h *WikiCQRSHandler) GetAnalyticsSummary(ctx context.Context, req *pb.GetAnalyticsSummaryRequest) (*pb.GetAnalyticsSummaryResponse, error) {
	days := int(req.Days)
	if days <= 0 {
		days = 30
	}
	totalViews, todayViews, dailyViews, pageRanking, err := h.analyticsRepo.GetSummary(ctx, days)
	if err != nil {
		slog.Error("failed to get analytics summary", "error", err)
		return nil, status.Error(codes.Internal, "failed to get analytics summary")
	}

	var pbDailyViews []*pb.DailyCount
	for _, dc := range dailyViews {
		pbDailyViews = append(pbDailyViews, &pb.DailyCount{
			Date:  dc.Date,
			Count: dc.Count,
		})
	}

	var pbPageRanking []*pb.PageRanking
	for _, pr := range pageRanking {
		pbPageRanking = append(pbPageRanking, &pb.PageRanking{
			Path:  pr.Path,
			Count: pr.Count,
		})
	}

	return &pb.GetAnalyticsSummaryResponse{
		TotalViews:  totalViews,
		TodayViews:  todayViews,
		DailyViews:  pbDailyViews,
		PageRanking: pbPageRanking,
	}, nil
}
