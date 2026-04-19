package config

import "os"

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
	GRPCPort   string
}

func Load() *Config {
	return &Config{
		DBUser:     os.Getenv("MYSQL_USER"),
		DBPassword: os.Getenv("MYSQL_PASSWORD"),
		DBHost:     getEnv("MYSQL_HOST", "localhost"),
		DBPort:     getEnv("MYSQL_PORT", "3306"),
		DBName:     os.Getenv("MYSQL_DATABASE"),
		GRPCPort:   getEnv("GRPC_PORT", "50052"),
	}
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
