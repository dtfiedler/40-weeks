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
	// Email configuration
	EmailEnabled    bool
	AWSRegion       string
	AWSAccessKeyID  string
	AWSSecretKey    string
	SenderEmail     string
	SenderName      string
}

var AppConfig *Config

func init() {
	AppConfig = &Config{
		JWTSecret:       getEnvWithDefault("JWT_SECRET", "your-secret-key-change-this"),
		ServerPort:      getEnvWithDefault("PORT", "8080"),
		DatabaseURL:     getEnvWithDefault("DATABASE_URL", "./data/sqlite/core.db"),
		ImagesDirectory: getEnvWithDefault("IMAGES_DIRECTORY", "./data/images"),
		VideosDirectory: getEnvWithDefault("VIDEOS_DIRECTORY", "./data/videos"),
		// Email configuration
		EmailEnabled:    GetEnvAsBool("EMAIL_ENABLED", false),
		AWSRegion:       getEnvWithDefault("AWS_REGION", "us-east-1"),
		AWSAccessKeyID:  getEnvWithDefault("AWS_ACCESS_KEY_ID", ""),
		AWSSecretKey:    getEnvWithDefault("AWS_SECRET_ACCESS_KEY", ""),
		SenderEmail:     getEnvWithDefault("SENDER_EMAIL", "noreply@40weeks.app"),
		SenderName:      getEnvWithDefault("SENDER_NAME", "40Weeks"),
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