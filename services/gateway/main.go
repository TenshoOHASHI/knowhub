package main

import (
	"log"
	"net/http"

	"github.com/TenshoOHASHI/knowhub/services/gateway/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	authConn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to wiki service %v", err)
	}
	defer authConn.Close()

	// クライアント側の接続を用意(TCPはfalseにする＝デジタル証明書)
	wikiConn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to wiki service %v", err)
	}
	defer wikiConn.Close()

	profileConn, err := grpc.NewClient("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to wiki service %v", err)
	}
	defer authConn.Close()

	// gRPCのハンドラーを作成
	wikiHandler := handler.NewWikiHandler(wikiConn)
	authHandler := handler.NewAuthHandler(authConn)
	profileHandler := handler.NewProfileHandle(profileConn)

	// ルーティング
	mux := http.NewServeMux()

	//　wiki
	mux.HandleFunc("GET /api/articles", wikiHandler.ListArticles)
	mux.HandleFunc("GET /api/articles/{id}", wikiHandler.GetArticle)
	mux.HandleFunc("POST /api/articles", wikiHandler.CreateArticle)
	mux.HandleFunc("PUT /api/articles/{id}", wikiHandler.UpdateArticle)
	mux.HandleFunc("DELETE /api/articles/{id}", wikiHandler.DeleteArticle)

	// auth
	mux.HandleFunc("POST /api/user/register", authHandler.Register)
	mux.HandleFunc("POST /api/user/login", authHandler.Login)

	// profile
	mux.HandleFunc("GET /api/profile", profileHandler.GetProfile)
	mux.HandleFunc("POST /api/profile", profileHandler.CreateProfile)
	mux.HandleFunc("PUT /api/profile", profileHandler.UpdateProfile)

	// item
	mux.HandleFunc("GET /api/portfolio", profileHandler.ListPortfolioItem)
	mux.HandleFunc("GET /api/portfolio/{id}", profileHandler.GetPortfolioItem)
	mux.HandleFunc("POST /api/portfolio", profileHandler.CreatePortfolioItem)
	mux.HandleFunc("PUT /api/portfolio/{id}", profileHandler.UpdatePortfolioItem)
	mux.HandleFunc("DELETE /api/portfolio/{id}", profileHandler.DeletePortfolioItem)

	log.Println("API Gateway started on: 8080")
	log.Fatal(http.ListenAndServe(":8080", mux)) // ルーティングを登録
}
