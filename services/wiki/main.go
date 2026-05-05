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

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"

	pb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/config"
	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/handler"
	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/repository"
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
	logger := loggerpkg.New("Wiki", cfg.LogLevel)
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

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisHost + ":" + cfg.RedisPort,
	})

	// CQRS: Command と Query を作成
	commandRepo := repository.NewMysqlCommandRepository(rdb, loggedDB)
	queryRepo := repository.NewMysqlQueryRepository(rdb, loggedDB)
	categoryRepo := repository.NewMysqlCategoryRepository(loggedDB)
	likeRepo := repository.NewMysqlLikeRepository(loggedDB)
	savedArticleRepo := repository.NewMysqlSavedArticleRepository(loggedDB)
	analyticsRepo := repository.NewMysqlAnalyticsRepository(loggedDB)

	// handlerを生成
	wikiCQRSHandler := handler.NewWikiCQRSHandler(commandRepo, queryRepo, categoryRepo, likeRepo, savedArticleRepo, analyticsRepo)

	// gRPC server
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	pb.RegisterWikiServicesServer(s, wikiCQRSHandler)
	reflection.Register(s)

	// goroutine でサーバー起動
	go func() {
		slog.Info("Wiki Service started", "port", cfg.GRPCPort)
		if err := s.Serve(lis); err != nil {
			slog.Error("failed to serve", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful Shutdown
	server.WaitForShutdown(s, db)
}
