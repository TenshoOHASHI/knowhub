package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	aiPb "github.com/TenshoOHASHI/knowhub/proto/ai"
	pb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"github.com/TenshoOHASHI/knowhub/services/mcp/internal/config"
	"github.com/TenshoOHASHI/knowhub/services/mcp/internal/handler"
	"github.com/mark3labs/mcp-go/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := config.Load("../../.env")

	// gRPC connections
	wikiConn := dialService("wiki", cfg.WikiAddr)
	aiConn := dialService("ai", cfg.AIAddr)
	defer wikiConn.Close()
	defer aiConn.Close()

	wikiClient := pb.NewWikiServicesClient(wikiConn)
	aiClient := aiPb.NewAIServiceClient(aiConn)

	mcpHandler := handler.NewMCPServer(wikiClient, aiClient)
	stdioServer := server.NewStdioServer(mcpHandler.GetServer())

	// Setup context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("shutting down MCP server")
		cancel()
	}()

	slog.Info("Starting TenHub MCP Server (stdio)")
	if err := stdioServer.Listen(ctx, os.Stdin, os.Stdout); err != nil {
		slog.Error("MCP server error", "error", err)
		os.Exit(1)
	}
}

func dialService(name, addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect to service", "service", name, "error", err)
		os.Exit(1)
	}
	conn.Connect()
	slog.Info("connected to service", "service", name, "addr", addr, "state", conn.GetState())
	return conn
}
