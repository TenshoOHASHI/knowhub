package logger

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

func New(serviceName string, logLevel string) *slog.Logger {
	// logs ディレクトリを自動作成
	os.MkdirAll("logs", 0755)

	// lumberjack
	writer := &lumberjack.Logger{
		Filename:   "logs/" + serviceName + ".log",
		MaxSize:    10,   // MB単位でローテンション
		MaxBackups: 5,    // 古いログファイルの保存数
		MaxAge:     30,   // 日数で自動削除
		Compress:   true, // 古いファイルをgzip圧縮
	}

	// コンソール+ファイル両方出力
	multiWriter := io.MultiWriter(os.Stdout, writer)

	level := parseLevel(logLevel)

	// ハンドラーのログを作成
	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level:     level,
		AddSource: true, // エラー時にソースファイル・行番号を記録
	})

	// インスタンスを作成（サービス名をデフォルト属性に追加）
	return slog.New(handler).With("service", serviceName)
}

// レベルごとに数値を返却
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
