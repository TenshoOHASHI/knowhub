package handler

import (
	"context"
	"database/sql"
	"fmt"

	pb "github.com/TenshoOHASHI/knowhub/proto/profile" // 生成されたprotoコード
	"github.com/TenshoOHASHI/knowhub/services/profile/internal/model"
	"github.com/TenshoOHASHI/knowhub/services/profile/internal/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProfileHandler はgRPCのハンドラー
type ProfileHandler struct {
	pb.UnimplementedProfileServiceServer                              // protoから生成された基底構造
	profileRepo                          repository.ProfileRepository // interface（テスト時にモック差し替え可）
	portfolioItemRepo                    repository.PortfolioItemRepository
}

func NewProfileHandler(profileRepo repository.ProfileRepository, portfolioItemRepo repository.PortfolioItemRepository) *ProfileHandler {
	return &ProfileHandler{
		profileRepo:       profileRepo,
		portfolioItemRepo: portfolioItemRepo,
	}
}

func (h *ProfileHandler) CreateProfile(ctx context.Context, req *pb.CreateProfileRequest) (*pb.CreateProfileResponse, error) {
	_, err := h.profileRepo.FindFirst(ctx)
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, "profile already exists")
	}

	if err != sql.ErrNoRows {
		return nil, status.Error(codes.Internal, "failed to check profile")
	}

	profile, err := model.NewProfile(req.Title, req.Bio, req.GithubUrl)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = h.profileRepo.Create(ctx, profile)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create profile: %v", err))
		// return nil, status.Error(codes.Internal, "failed to create profile")
	}
	return &pb.CreateProfileResponse{
		Profile: toProfile(profile),
	}, nil
}

func (h *ProfileHandler) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	profile, err := h.profileRepo.FindFirst(ctx)
	if err != nil {

		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		return nil, status.Error(codes.Internal, "failed to find profile")
	}

	return &pb.GetProfileResponse{
		Profile: toProfile(profile),
	}, nil

}

func (h *ProfileHandler) GetPortfolioItem(ctx context.Context, req *pb.GetPortfolioItemRequest) (*pb.GetPortfolioItemResponse, error) {
	item, err := h.portfolioItemRepo.FindById(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "item not found")
		}
		return nil, status.Error(codes.Internal, "failed to get item")
	}
	return &pb.GetPortfolioItemResponse{
		Item: toPortfolioItem(item),
	}, nil
}

func (h *ProfileHandler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	profile, err := h.profileRepo.FindFirst(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		return nil, status.Error(codes.Internal, "failed to find profile")
	}

	profile.Update(req.Title, req.Bio, req.GithubUrl)
	err = h.profileRepo.Save(ctx, profile)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update profile")
	}
	return &pb.UpdateProfileResponse{
		Profile: toProfile(profile),
	}, nil

}

func (h *ProfileHandler) CreatePortfolioItem(ctx context.Context, req *pb.CreatePortfolioItemRequest) (*pb.CreatePortfolioItemResponse, error) {

	item, err := model.NewPortfolioItem(req.Title, req.Description, req.Url, model.PortfolioStatus(req.Status))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	err = h.portfolioItemRepo.Create(ctx, item)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to save item")
	}

	return &pb.CreatePortfolioItemResponse{
		Item: toPortfolioItem(item),
	}, nil
}

func (h *ProfileHandler) ListPortfolioItems(ctx context.Context, req *pb.ListPortfolioItemsRequest) (*pb.ListPortfolioItemsResponse, error) {
	items, err := h.portfolioItemRepo.FindAll(ctx)
	if err != nil {

		return nil, status.Error(codes.Internal, "failed to get items")
	}

	// 1つのレスポンスに対して、リストが入る
	return &pb.ListPortfolioItemsResponse{
		Items: toPortfolioItems(items),
	}, nil
}

func (h *ProfileHandler) UpdatePortfolioItem(ctx context.Context, req *pb.UpdatePortfolioItemRequest) (*pb.UpdatePortfolioItemResponse, error) {
	item, err := h.portfolioItemRepo.FindById(ctx, req.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "item not found")
		}
		return nil, status.Error(codes.Internal, "failed to get item")
	}

	var title, description, url string
	var portfolioStatus model.PortfolioStatus
	if req.Title != nil {
		title = *req.Title
	}
	if req.Description != nil {
		description = *req.Description
	}

	if req.Url != nil {
		url = *req.Url
	}

	if req.Status != nil {
		portfolioStatus = model.PortfolioStatus(*req.Status)
	}

	item.Update(title, description, url, portfolioStatus)
	err = h.portfolioItemRepo.Save(ctx, item)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update portfolio item")
	}

	return &pb.UpdatePortfolioItemResponse{
		Item: toPortfolioItem(item),
	}, nil
}

func (h *ProfileHandler) DeletePortfolioItem(ctx context.Context, req *pb.DeletePortfolioItemRequest) (*emptypb.Empty, error) {
	err := h.portfolioItemRepo.Delete(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete item")
	}
	return &emptypb.Empty{}, nil
}

func toProfile(p *model.Profile) *pb.Profile {
	return &pb.Profile{
		Id:        p.ID,
		Title:     p.Title,
		Bio:       p.Bio,
		GithubUrl: p.GithubURL,
		CreatedAt: timestamppb.New(p.CreatedAt),
		UpdatedAt: timestamppb.New(p.UpdatedAt),
	}
}

func toPortfolioItem(item *model.PortfolioItem) *pb.PortfolioItem {
	return &pb.PortfolioItem{
		Id:          item.ID,
		Title:       item.Title,
		Description: item.Description,
		Url:         item.URL,
		Status:      string(item.Status),
		CreatedAt:   timestamppb.New(item.CreatedAt),
	}
}

func toPortfolioItems(items []*model.PortfolioItem) []*pb.PortfolioItem {
	var result []*pb.PortfolioItem
	for _, item := range items {
		result = append(result, toPortfolioItem(item))
	}
	return result
}
