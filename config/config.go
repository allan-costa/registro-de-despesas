package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	ServerPort  string
	CORSOrigin  string
}

// Load lê o .env (opcional) e aplica defaults; sem JWT_SECRET gera um aleatório (tokens não sobrevivem a reinício).
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres@localhost:5432/registro_despesas?sslmode=disable"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		ServerPort:  getEnv("SERVER_PORT", "5000"),
		CORSOrigin:  getEnv("CORS_ORIGIN", "*"),
	}

	if cfg.JWTSecret == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return nil, fmt.Errorf("gerar JWT secret: %w", err)
		}
		cfg.JWTSecret = hex.EncodeToString(b)
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
