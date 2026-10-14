package middleware

import (
	"log/slog"

	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
)

// middleware de log para registrar cada requisição feita a api
func Logger(log *slog.Logger) echo.MiddlewareFunc {
	return echomw.RequestLoggerWithConfig(echomw.RequestLoggerConfig{
		LogMethod:   true,
		LogURI:      true,
		LogStatus:   true,
		LogLatency:  true,
		LogRemoteIP: true,
		HandleError: true,
		LogValuesFunc: func(c *echo.Context, v echomw.RequestLoggerValues) error {
			attrs := []any{
				"method", v.Method,
				"uri", v.URI,
				"status", v.Status,
				"latency", v.Latency.String(),
				"ip", v.RemoteIP,
			}
			if v.Error != nil {
				log.Error("request", append(attrs, "error", v.Error)...)
			} else {
				log.Info("request", attrs...)
			}
			return nil
		},
	})
}
