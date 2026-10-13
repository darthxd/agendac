package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env            string
	Port           string
	DBDriver       string
	DBDSN          string
	JWTSecret      string
	JWTExpiration  time.Duration
	AllowedOrigins []string

	AdminName     string
	AdminEmail    string
	AdminPassword string
}

// carrega os dados do .env para uma struct Config
func Load() (*Config, error) {
	_ = godotenv.Load()

	// cria uma struct Config e salva seu endereço de memória na variável "cfg"
	cfg := &Config{
		Env:            strings.ToLower(getEnv("APP_ENV", "development")),
		Port:           getEnv("PORT", "8080"),
		DBDriver:       strings.ToLower(getEnv("DB_DRIVER", "sqlite")),
		DBDSN:          getEnv("DB_DSN", "app.db"),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		AllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "*")),
		AdminName:      getEnv("ADMIN_NAME", "admin"),
		AdminEmail:     getEnv("ADMIN_EMAIL", ""),
		AdminPassword:  getEnv("ADMIN_PASSWORD", ""),
	}

	exp, err := time.ParseDuration(getEnv("JWT_EXPIRATION", "24h"))
	if err != nil {
		return nil, fmt.Errorf("JWT_EXPIRATION inválido: %w", err)
	}
	cfg.JWTExpiration = exp

	if cfg.JWTSecret == "" {
		if cfg.Env == "production" {
			return nil, fmt.Errorf("JWT_SECRET é obrigatório em produção")
		}
		cfg.JWTSecret = "pJ9ovuEGVN3JO8soaYNjQZiAGQcPAwJ7NLa7Hn7JRqs"
	}

	return cfg, nil
}

// função helper para buscar valores do .env, ou retornar valores pré-definidos caso não encontre
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

// função helper para splitar uma string por vírgula
func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	output := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			output = append(output, p)
		}
	}
	return output
}
