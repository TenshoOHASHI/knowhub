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
	defer profileConn.Close()

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

	// categories
	mux.HandleFunc("GET /api/categories", wikiHandler.ListCategories)
	mux.HandleFunc("POST /api/categories", wikiHandler.CreateCategory)
	mux.HandleFunc("DELETE /api/categories/{id}", wikiHandler.DeleteCategory)

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
	log.Fatal(http.ListenAndServe(":8080", corsMiddleware(mux))) // ルーティングを登録
}

// middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 許可するオリジン
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		// 許可するメソッド
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// 許可するヘッダー
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// プリフライトリクエスト（OPTIONS）は即レスポンス
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
