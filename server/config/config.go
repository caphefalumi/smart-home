package config

import "os"

// Config holds application configuration
type Config struct {
	Port     string
	MongoURI string
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Port:     getEnv("PORT", "3000"),
		MongoURI: getEnv("MONGODB_URI", "mongodb://localhost:27017/smarthome"),
	}
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
