package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	pb "github.com/TenshoOHASHI/knowhub/proto/profile" // 生成されたprotoコード

	"github.com/TenshoOHASHI/knowhub/services/profile/internal/config"
	"github.com/TenshoOHASHI/knowhub/services/profile/internal/handler"
	"github.com/TenshoOHASHI/knowhub/services/profile/internal/repository"
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load("../../.env")
	dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	// mysqlに接続
	db, err := sql.Open("mysql", dns)

	if err != nil {
		// 終了
		panic(err)
	}

	defer db.Close()

	profileRepo := repository.NewMysqlProfileRepository(db)
	portfolioRepo := repository.NewMysqlPortfolioItemRepository(db)
	profileHandler := handler.NewProfileHandler(profileRepo, portfolioRepo)

	// grpcサーバを起動
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterProfileServiceServer(s, profileHandler)
	reflection.Register(s)

	log.Printf("Profile Service started on :%s", cfg.GRPCPort)
	// サーバ開始（ブロックする）
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
