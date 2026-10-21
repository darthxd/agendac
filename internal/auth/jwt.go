package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
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

// gera o token jwt
func GenerateToken(id uint) (string, error) {
	claims := &jwt.MapClaims{
		"user_id": id,
		"exp":     jwt.NewNumericDate(time.Now().Add(*expiration)),
		"iat":     jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(*secret))
}

// recupera as claims do token jwt
func ParseClaims(c *echo.Context) (jwt.MapClaims, error) {
	token, err := echo.ContextGet[*jwt.Token](c, "user")
	if err != nil {
		return nil, err
	}
	return token.Claims.(jwt.MapClaims), nil
}
