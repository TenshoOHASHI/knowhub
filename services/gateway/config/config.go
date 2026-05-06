package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// CORS
	AllowedOrigin     string
	AllowedMethods    string
	AllowedHeaders    string
	AllowedCredential string

	// Services
	AuthAddr    string
	WikiAddr    string
	ProfileAddr string
	AIAddr      string

	// Server
	Port     string
	LogLevel string

	// Upload
	UploadDir string

	// Slack
	SlackWebhookURL string

	// AI rate limit
	AIAnonMaxConcurrent int
	AIAnonDailyLimit    int
}

func Load(envPath string) *Config {
	godotenv.Load(envPath)
	return &Config{
		AllowedOrigin:       getEnv("ALLOWED_ORIGIN", "http://localhost:3000"),
		AllowedMethods:      getEnv("ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"),
		AllowedHeaders:      getEnv("ALLOWED_HEADERS", "Content-Type,Authorization"),
		AllowedCredential:   getEnv("ALLOWED_CREDENTIALS", "true"),
		AuthAddr:            getEnv("AUTH_ADDR", "localhost:50051"),
		WikiAddr:            getEnv("WIKI_ADDR", "localhost:50052"),
		ProfileAddr:         getEnv("PROFILE_ADDR", "localhost:50053"),
		AIAddr:              getEnv("AI_ADDR", "localhost:50054"),
		Port:                getEnv("GATEWAY_PORT", "8080"),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		UploadDir:           getEnv("UPLOAD_DIR", "./uploads"),
		SlackWebhookURL:     getEnv("SLACK_WEBHOOK_URL", ""),
		AIAnonMaxConcurrent: getEnvInt("AI_ANON_MAX_CONCURRENT", 1),
		AIAnonDailyLimit:    getEnvInt("AI_ANON_DAILY_LIMIT", 5),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
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
