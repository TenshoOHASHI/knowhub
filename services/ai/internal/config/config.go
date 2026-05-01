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

	// LLM
	LLMProvider       string // "ollama", "glm5", "openai"
	EmbeddingProvider string // "ollama", "openai", "deepseek", "glm"
	EmbeddingModel    string // "nomic-embed-text", "text-embedding-3-small" 等

	OllamaURL     string
	OllamaModel   string // "qwen3:1.7b", "gemma3:1b"
	GLM5APIKey    string
	GLM5Model     string
	OpenAIKey     string
	GeminiKey     string
	GeminiModel   string
	DeepSeekKey   string
	DeepSeekModel string

	// Logging
	LogLevel string

	SearchEngin string // "BM25", "tfidf", "vector"
}

func Load(path string) *Config {
	godotenv.Load(path)
	return &Config{
		GRPCPort:          getEnv("GRPC_PORT_AI", "50054"),
		WikiAddr:          getEnv("WIKI_ADDR", "localhost:50052"),
		LLMProvider:       getEnv("LLM_PROVIDER", "ollama"),
		OllamaURL:         getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:       getEnv("OLLAMA_MODEL", "gemma3:1b"),
		GLM5APIKey:        os.Getenv("GLM5_API_KEY"),
		GLM5Model:         getEnv("GLM5_MODEL", "glm-5"),
		OpenAIKey:         os.Getenv("OPENAI_API_KEY"),
		GeminiKey:         os.Getenv("GEMINI_API_KEY"),
		GeminiModel:       getEnv("GEMINI_MODEL", "gemini-3-flash-preview"),
		DeepSeekKey:       os.Getenv("DEEPSEEK_API_KEY"),
		DeepSeekModel:     getEnv("DEEPSEEK_MODEL", "deepseek-v4-flash"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		SearchEngin:       getEnv("SEARCH_ENGIN", "BM25"),
		EmbeddingProvider: getEnv("EMBEDDING_PROVIDER", "ollama"),
		EmbeddingModel:    getEnv("EMBEDDING_MODEL", "nomic-embed-text"),
	}
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
