package config

import (
	"io"
	"log/slog"
	"os"

	"github.com/natefinch/lumberjack"
)

func setUpLogger() *slog.Logger {
	// lumberjack
	writer := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10,   // MB単位でローテンション
		MaxBackups: 5,    // 古いログファイルの保存数
		MaxAge:     30,   // 日数で自動削除
		Compress:   true, // 古いファイルをgzip圧縮
	}
	// コンソール+ファイル両方出力
	multiWriter := io.MultiWriter(os.Stdout, writer)

	// ハンドラーのログを作成
	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	// インスタンスを作成
	return slog.New(handler)
}
