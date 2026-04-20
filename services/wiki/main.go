package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // init関数だけ使う（database/sqlに登録する）
	"github.com/redis/go-redis/v9"

	"log"
	"net"

	pb "github.com/TenshoOHASHI/knowhub/proto/wiki" // 生成されたprotoコード
	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/config"
	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/handler"
	"github.com/TenshoOHASHI/knowhub/services/wiki/internal/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 初期設定
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

	// redisに接続
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisHost + ":" + cfg.RedisPort,
	})

	defer rdb.Close()

	// CQRS: Command と Query を作成
	// repo := repository.NewMysqlRepository(db)
	commandRepo := repository.NewMysqlCommandRepository(db)
	queryRepo := repository.NewMysqlQueryRepository(rdb, db)

	//  handlerを生成
	// wikiHandler := handler.NewWikiHandler(repo)
	wikiCQRSHandler := handler.NewWikiCQRSHandler(commandRepo, queryRepo)

	// grpcサーバを起動
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWikiServicesServer(s, wikiCQRSHandler)
	reflection.Register(s)

	log.Printf("Wiki Service started on :%s", cfg.GRPCPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	} // サーバ開始（ブロックする）
}
