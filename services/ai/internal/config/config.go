package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// gRPC
	GRPCPort string

	// Wiki Service（gRPC client）
	WikiAddr string

	// Ollama（デフォルト LLM + Embedding）
	OllamaURL      string
	OllamaModel    string // "gemma3:1b"
	EmbeddingModel string // "nomic-embed-text"

	// SearXNG（外部 Web 検索）
	SearXNGURL string

	// Logging
	LogLevel string
}

func Load(path string) *Config {
	godotenv.Load(path)
	return &Config{
		GRPCPort:       getEnv("GRPC_PORT_AI", "50054"),
		WikiAddr:       getEnv("WIKI_ADDR", "localhost:50052"),
		OllamaURL:      getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:    getEnv("OLLAMA_MODEL", "gemma3:1b"),
		EmbeddingModel: getEnv("EMBEDDING_MODEL", "nomic-embed-text"),
		SearXNGURL:     getEnv("SEARXNG_URL", "http://localhost:8888"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
