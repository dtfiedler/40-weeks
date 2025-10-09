package config

import (
	"os"
	"strconv"
)

type Config struct {
	JWTSecret       string
	ServerPort      string
	DatabaseURL     string
	ImagesDirectory string
	VideosDirectory string
}

var AppConfig *Config

func init() {
	AppConfig = &Config{
		JWTSecret:       getEnvWithDefault("JWT_SECRET", "your-secret-key-change-this"),
		ServerPort:      getEnvWithDefault("PORT", "8080"),
		DatabaseURL:     getEnvWithDefault("DATABASE_URL", "./data/sqlite/core.db"),
		ImagesDirectory: getEnvWithDefault("IMAGES_DIRECTORY", "./data/images"),
		VideosDirectory: getEnvWithDefault("VIDEOS_DIRECTORY", "./data/videos"),
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func GetEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}