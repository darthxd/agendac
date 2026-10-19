package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtConfig struct {
	JwtSecret     string
	JwtExpiration time.Duration
}

var (
	secret     *string
	expiration *time.Duration
)

func WithConfig(cfg JwtConfig) {
	secret = &cfg.JwtSecret
	expiration = &cfg.JwtExpiration
}

func GenerateToken(id uint, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  id,
		"exp":  time.Now().Add(*expiration).Unix(),
		"iat":  time.Now().Unix(),
		"role": role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(*secret))
}
