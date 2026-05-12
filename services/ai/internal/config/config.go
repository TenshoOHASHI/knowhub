package config

import (
	"os"
	"strconv"

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

	// DeepSeek（サーバー側で無料提供）
	DeepSeekAPIKey   string
	DeepSeekModel    string
	DeepSeekMaxTokens int

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
		SearXNGURL:      getEnv("SEARXNG_URL", "http://localhost:8888"),
		DeepSeekAPIKey:  getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekModel:   getEnv("DEEPSEEK_MODEL", "deepseek-v4-flash"),
		DeepSeekMaxTokens: getEnvInt("DEEPSEEK_MAX_TOKENS", 1000),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n >= 0 {
			return n
		}
	}
	return fallback
}
