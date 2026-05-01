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

	// LLM Provider（デフォルト: Ollama、リクエスト毎に handler で動的切替）
	provider := llm.NewOllamaProvider(cfg.OllamaURL, cfg.OllamaModel)

	// Search Engine（デフォルト: BM25、リクエスト毎に handler で動的切替）
	searchEngine := search.NewBM25Engine()

	// Handler
	aiHandler := handler.NewAIHandler(searchEngine, provider, cfg.OllamaURL, cfg.EmbeddingModel, wikiClient)

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
