package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
	JWTSecret   string
}

// Load reads config from .env (if present) and the real environment,
// the latter always wins so this works both locally and in prod/docker.
func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment variables")
	}
	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	// DB and JWT secret are required — better to fail fast on startup
	// than to hit a nil pool or an empty signing key mid-request
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	return cfg
}
