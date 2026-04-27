package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
	GRPCPort   string
	LogLevel   string
}

func Load(path string) *Config {
	godotenv.Load(path)
	return &Config{
		DBUser:     os.Getenv("MYSQL_USER"),
		DBPassword: os.Getenv("MYSQL_PASSWORD"),
		DBHost:     getEnv("MYSQL_HOST", "localhost"),
		DBPort:     getEnv("MYSQL_PORT", "3306"),
		DBName:     os.Getenv("MYSQL_DATABASE"),
		GRPCPort:   getEnv("GRPC_PORT_AUTH", "50051"),
		LogLevel:   os.Getenv("LOG_LEVEL"),
	}
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
