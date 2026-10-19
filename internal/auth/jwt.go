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

func Use(cfg JwtConfig) {
	secret = &cfg.JwtSecret
	expiration = &cfg.JwtExpiration
}

func GenerateToken(id uint, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  id,
		"exp":  time.Now().Add(*expiration),
		"iat":  time.Now(),
		"role": role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(*secret))
}
