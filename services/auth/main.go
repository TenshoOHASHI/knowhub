package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	pb "github.com/TenshoOHASHI/knowhub/proto/auth" // 生成されたprotoコード
	"github.com/TenshoOHASHI/knowhub/services/auth/internal/config"
	"github.com/TenshoOHASHI/knowhub/services/auth/internal/handler"
	"github.com/TenshoOHASHI/knowhub/services/auth/internal/repository"
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

	db, err := sql.Open("mysql", dns)

	if err != nil {
		// 終了
		panic(err)
	}

	defer db.Close()

	// create repo
	repo := repository.NewMysqlRepository(db)
	// create handler
	authHandler := handler.NewUserHandler(repo)
	// grpc server
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, authHandler)
	reflection.Register(s)

	log.Printf("Auth Service started on :%s", cfg.GRPCPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	} // サーバ開始（ブロックする）
}
