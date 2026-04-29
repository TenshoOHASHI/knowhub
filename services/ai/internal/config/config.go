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
	LLMProvider string // "ollama", "glm5", "openai"
	OllamaURL   string
	GLM5APIKey  string
	GLM5Model   string
	OpenAIKey   string

	// Logging
	LogLevel string

	SearchEngin string
}

func Load(path string) *Config {
	godotenv.Load(path)
	return &Config{
		GRPCPort:    getEnv("GRPC_PORT_AI", "50054"),
		WikiAddr:    getEnv("WIKI_ADDR", "localhost:50052"),
		LLMProvider: getEnv("LLM_PROVIDER", "ollama"),
		OllamaURL:   getEnv("OLLAMA_URL", "http://localhost:11434"),
		GLM5APIKey:  os.Getenv("GLM5_API_KEY"),
		GLM5Model:   getEnv("GLM5_MODEL", "glm-5"),
		OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		SearchEngin: os.Getenv("SEARCH_ENGIN"),
	}
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
