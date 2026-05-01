package main

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	loggerpkg "github.com/TenshoOHASHI/knowhub/pkg/logger"

	pb "github.com/TenshoOHASHI/knowhub/proto/auth"
	aiPb "github.com/TenshoOHASHI/knowhub/proto/ai"

	"github.com/TenshoOHASHI/knowhub/services/gateway/config"
	"github.com/TenshoOHASHI/knowhub/services/gateway/handler"
	"github.com/TenshoOHASHI/knowhub/services/gateway/middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/TenshoOHASHI/knowhub/services/gateway/docs" // 自動生成先
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title TenHub API Gateway
// @version 1.0
// @description TenHub テクニカルナレッジベースプラットフォームのAPI
// @host localhost:8080
// @BasePath /
// @schemas http
func main() {
	// config
	cfg := config.Load("../../.env")
	// Logger
	logger := loggerpkg.New("Gateway", cfg.LogLevel)
	slog.SetDefault(logger)

	// gRPC connections
	authConn := dialService("auth", cfg.AuthAddr)
	wikiConn := dialService("wiki", cfg.WikiAddr)
	profileConn := dialService("profile", cfg.ProfileAddr)
	aiConn := dialService("ai", cfg.AIAddr)

	// close connection
	defer authConn.Close()
	defer wikiConn.Close()
	defer profileConn.Close()
	defer aiConn.Close()

	// Handlers
	wikiHandler := handler.NewWikiHandler(wikiConn)
	wikiHandler.SetAIClient(aiPb.NewAIServiceClient(aiConn))
	authHandler := handler.NewAuthHandler(authConn)
	profileHandler := handler.NewProfileHandle(profileConn)
	uploadHandler := handler.NewUploadHandler(cfg.UploadDir)
	aiHandler := handler.NewAIHandler(aiConn)

	// Routes
	mux := http.NewServeMux()

	// wiki
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
	mux.HandleFunc("GET /api/user/me", authHandler.Me)

	// profile
	mux.HandleFunc("GET /api/profile", profileHandler.GetProfile)
	mux.HandleFunc("POST /api/profile", profileHandler.CreateProfile)
	mux.HandleFunc("PUT /api/profile", profileHandler.UpdateProfile)

	// portfolio
	mux.HandleFunc("GET /api/portfolio", profileHandler.ListPortfolioItem)
	mux.HandleFunc("GET /api/portfolio/{id}", profileHandler.GetPortfolioItem)
	mux.HandleFunc("POST /api/portfolio", profileHandler.CreatePortfolioItem)
	mux.HandleFunc("PUT /api/portfolio/{id}", profileHandler.UpdatePortfolioItem)
	mux.HandleFunc("DELETE /api/portfolio/{id}", profileHandler.DeletePortfolioItem)

	// ai
	mux.HandleFunc("POST /api/ai/search", aiHandler.SearchArticles)
	mux.HandleFunc("POST /api/ai/summarize", aiHandler.SummarizeArticle)
	mux.HandleFunc("POST /api/ai/ask", aiHandler.AskQuestion)
	mux.HandleFunc("GET /api/ai/graph", aiHandler.GetKnowledgeGraph)

	// upload
	mux.HandleFunc("POST /api/upload", uploadHandler.Upload)

	// static file serving for uploads (development)
	// ルートディレクトの指定、静的ファイルサーバ、リクエストされたパスのファイルを読み込んでレスポンスとして返す、また、プレフィックスを削除
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir))))

	// Swagger(*スラッシュ付き)
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("doc.json"),
	))

	// CORS -> Auth -> Routing
	// Ensure upload directory exists
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		slog.Error("failed to create upload directory", "error", err)
		os.Exit(1)
	}

	authMW := middleware.NewAuthMiddleware(pb.NewAuthServiceClient(authConn))
	handler := middleware.NewCoreMiddleware(cfg.AllowedOrigin, cfg.AllowedMethods, cfg.AllowedHeaders, cfg.AllowedCredential).CorsMiddleware(authMW.RequireAuth(mux))

	// goroutine でサーバー起動
	go func() {
		slog.Info("API Gateway started", "port", cfg.Port)
		if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
			slog.Error("failed to serve", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful Shutdown（Gateway は DB なし）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("shutting down", "signal", sig)
	slog.Info("server stopped gracefully")
}

func dialService(serviceName, addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect to service", "service", serviceName, "error", err)
		os.Exit(1)
	}

	conn.Connect()
	state := conn.GetState()
	slog.Info("connected to service", "service", serviceName, "addr", addr, "state", state)

	return conn
}
