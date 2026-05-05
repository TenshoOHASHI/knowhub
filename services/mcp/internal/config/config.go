package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	WikiAddr string
	AIAddr   string
}

func Load(envPath string) *Config {
	godotenv.Load(envPath)
	return &Config{
		WikiAddr: getEnv("WIKI_ADDR", "localhost:50052"),
		AIAddr:   getEnv("AI_ADDR", "localhost:50054"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
