package server

import (
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
)

func WaitForShutdown(grpcServer *grpc.Server, db *sql.DB) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // ctrl+c or q

	sig := <-quit
	grpcServer.GracefulStop()

	if db != nil {
		db.Close()
	}

	slog.Info("server stopped gracefully", "signal", sig)
}
