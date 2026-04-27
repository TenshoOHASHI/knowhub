package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"os"

	loggerpkg "github.com/TenshoOHASHI/knowhub/pkg/logger"
	"github.com/TenshoOHASHI/knowhub/pkg/server"

	"github.com/TenshoOHASHI/knowhub/pkg/dbutil"
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

	// Logger
	logger := loggerpkg.New("Auth", cfg.LogLevel)
	slog.SetDefault(logger)

	// DB
	db, err := sql.Open("mysql", dns)
	if err != nil {
		// 終了
		slog.Error("failed to connect db", "error", err)
		os.Exit(1)

	}

	// DB生存確認
	if err := db.Ping(); err != nil {
		slog.Error("DB ping failed", "error", err)
		os.Exit(1)
	}

	slog.Info("DB connected")

	// DB をログ付きラッパーで包む
	loggedDB := dbutil.Wrap(db)

	// create repo amd handler
	repo := repository.NewMysqlRepository(loggedDB)
	authHandler := handler.NewUserHandler(repo)

	// gRPC server
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		slog.Error("failed to serve", "error", err)
		os.Exit(1)

	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, authHandler)
	reflection.Register(s)

	// goroutine でサーバー起動
	go func() {
		slog.Info("Auth Service started", "port", cfg.GRPCPort)
		if err := s.Serve(lis); err != nil {
			slog.Error("failed to serve", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful Shutdown
	server.WaitForShutdown(s, db)
}
