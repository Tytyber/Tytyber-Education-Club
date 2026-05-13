package config

import "os"

type Config struct {
	Port          string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPass        string
	DBName        string
	SessionSecret string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", ":8080"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPass:        getEnv("DB_PASS", ""),
		DBName:        getEnv("DB_NAME", "tec_db"),
		SessionSecret: getEnv("SESSION_SECRET", "change-me-in-prod"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
