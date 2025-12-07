package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      string
	JWTSecret string
	DBPath    string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	// Load the.env file if present (non-fatal if missing)
	_ = godotenv.Load()
	cfg := Config{
		Port:      getEnv("PORT", "8081"),
		JWTSecret: getEnv("JWT_SECRET", "change-me-in-env"),
		DBPath:    getEnv("DB_PATH", "app.db"),
	}
	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
