package handler

import (
	"context"
	"database/sql"
	"log"

	pb "github.com/TenshoOHASHI/knowhub/proto/auth"
	"github.com/TenshoOHASHI/knowhub/services/auth/internal/model"
	"github.com/TenshoOHASHI/knowhub/services/auth/internal/repository"
	"github.com/TenshoOHASHI/knowhub/services/auth/jwt"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/go-sql-driver/mysql"
)

type UserHandler struct {
	pb.UnimplementedAuthServiceServer                           // 基底クラスを埋め込む
	repo                              repository.UserRepository // 外部からrepoを受け取る
}

// repoを外部から注入する際に、UserRepositoryを満たす必要がある
// handlerはrepo(db)の中身を知らなくていい
func NewUserHandler(repo repository.UserRepository) *UserHandler {
	return &UserHandler{
		repo: repo,
	}
}

func (h *UserHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := model.NewUser(req.Username, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to register user")
	}

	// 作成・保存
	err = h.repo.Create(ctx, user)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return nil, status.Error(codes.AlreadyExists, "email already exists")
		}
		// return nil, status.Error(codes.Internal, "Failed to save user")
		log.Printf("failed to save user: %v", err) // ← サーバーログに詳細
		return nil, status.Error(codes.Internal, "Failed to save user: %")

	}

	token, err := jwt.GenerateToken(user.ID, req.Username)
	if err != nil {
		log.Printf("Failed to crate token: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create token")
	}
	return &pb.RegisterResponse{
		User:  toProtoUser(user),
		Token: token,
	}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// ドメインのバリデーションチェック
	user, err := h.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "username or password is correct")
	}

	// パスワード認証（新規で生成された、ユーザー情報と実際に渡ってきたパスワードと比較）
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid password")
	}
	// トークンを生成
	token, err := jwt.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	// pbの型でデータを返す（domain -> proto）
	return &pb.LoginResponse{
		User:  toProtoUser(user),
		Token: token,
	}, nil
}

func toProtoUser(u *model.User) *pb.User {
	return &pb.User{
		Id:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		CreateAt: timestamppb.New(u.CreatedAt), // ユーザーの作成日時
	}
}
