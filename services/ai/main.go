package main

import (
	"log/slog"
	"net"
	"os"

	loggerpkg "github.com/TenshoOHASHI/knowhub/pkg/logger"

	pb "github.com/TenshoOHASHI/knowhub/proto/ai"
	wikiPb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/config"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/handler"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/llm"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/search"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.Load("../../.env")

	// Logger
	logger := loggerpkg.New("AI", cfg.LogLevel)
	slog.SetDefault(logger)

	// Wiki Service への gRPC 接続
	wikiConn, err := grpc.NewClient(cfg.WikiAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect to wiki service", "error", err)
		os.Exit(1)
	}
	defer wikiConn.Close()
	wikiClient := wikiPb.NewWikiServicesClient(wikiConn)
	slog.Info("connected to wiki service", "addr", cfg.GRPCPort)

	// LLM Provider（環境変数で切り替え）
	var provider llm.LLMProvider
	switch cfg.LLMProvider {
	case "glm5":
		provider = llm.NewGLM5Provider(cfg.GLM5APIKey, cfg.GLM5Model)
		slog.Info("LLM provider: GLM-5")
	case "openai":
		provider = llm.NewOpenAIProvider(cfg.OpenAIKey)
		slog.Info("LLM provider: OpenAI")
	case "gemini":
		provider = llm.NewGeminiProvider(cfg.GeminiKey, cfg.GeminiModel)
		slog.Info("LLM provider: Gemini", "model", cfg.GeminiModel)
	case "deepseek":
		provider = llm.NewDeepSeekProvider(cfg.DeepSeekKey, cfg.DeepSeekModel)
		slog.Info("LLM provider: DeepSeek", "model", cfg.DeepSeekModel)
	default:
		provider = llm.NewOllamaProvider(cfg.OllamaURL, cfg.OllamaModel)
		slog.Info("LLM provider: Ollama", "url", cfg.OllamaURL)
	}

	var searchEngine search.SearchEngine

	switch cfg.SearchEngin {
	case "tfidf":
		searchEngine = search.NewTFIDFEngine()
		slog.Info("Search engine: TF-IDF")
	default:
		searchEngine = search.NewBM25Engine()
		slog.Info("Search engine: BM25")
	}

	// Handler
	aiHandler := handler.NewAIHandler(searchEngine, provider, wikiClient)

	// gRPC Server
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	pb.RegisterAIServiceServer(s, aiHandler)
	reflection.Register(s)

	go func() {
		slog.Info("AI Service started", "port", cfg.GRPCPort)
		if err := s.Serve(lis); err != nil {
			slog.Error("failed to serve", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful Shutdown（DB なし）
	quit := make(chan os.Signal, 1)
	// signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	s.GracefulStop()
	slog.Info("server stopped gracefully", "signal", sig)
}
