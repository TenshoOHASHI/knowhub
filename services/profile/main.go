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

	pb "github.com/TenshoOHASHI/knowhub/proto/profile"
	"github.com/TenshoOHASHI/knowhub/services/profile/internal/config"
	"github.com/TenshoOHASHI/knowhub/services/profile/internal/handler"
	"github.com/TenshoOHASHI/knowhub/services/profile/internal/repository"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Config
	cfg := config.Load("../../.env")
	dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	// Logger
	logger := loggerpkg.New("Profile", cfg.LogLevel)
	slog.SetDefault(logger)

	// MySQL
	db, err := sql.Open("mysql", dns)
	if err != nil {
		slog.Error("failed to connect DB", "error", err)
		os.Exit(1)
	}
	if err := db.Ping(); err != nil {
		slog.Error("DB ping failed", "error", err)
		os.Exit(1)
	}
	slog.Info("DB connected")

	// DB をログ付きラッパーで包む
	loggedDB := dbutil.Wrap(db)

	// Repository & Handler
	profileRepo := repository.NewMysqlProfileRepository(loggedDB)
	portfolioRepo := repository.NewMysqlPortfolioItemRepository(loggedDB)
	profileHandler := handler.NewProfileHandler(profileRepo, portfolioRepo)

	// gRPC server
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	pb.RegisterProfileServiceServer(s, profileHandler)
	reflection.Register(s)

	// goroutine でサーバー起動
	go func() {
		slog.Info("Profile Service started", "port", cfg.GRPCPort)
		if err := s.Serve(lis); err != nil {
			slog.Error("failed to serve", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful Shutdown
	server.WaitForShutdown(s, db)
}
