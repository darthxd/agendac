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
	claims := &jwt.MapClaims{
		"user_id": id,
		"exp":     jwt.NewNumericDate(time.Now().Add(*expiration)),
		"iat":     jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(*secret))
}
