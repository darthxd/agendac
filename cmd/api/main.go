package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/darthxd/agendac/internal/auth"
	"github.com/darthxd/agendac/internal/config"
	"github.com/darthxd/agendac/internal/database"
	"github.com/darthxd/agendac/internal/middleware"
	"github.com/darthxd/agendac/internal/user"
	"github.com/darthxd/agendac/pkg/logger"
	"github.com/labstack/echo/v5"
)

func main() {
	if err := run(); err != nil {
		slog.Error("aplicação encerrada com erro", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.Env)
	slog.SetDefault(log)

	db, err := database.Connect(cfg)
	if err != nil {
		return err
	}
	defer database.Close(db)

	if err := database.Migrate(db, &user.User{}); err != nil {
		return err
	}

	e := echo.New()
	e.Logger = log
	e.Use(middleware.Logger(log))

	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	auth.WithConfig(auth.JwtConfig{
		JwtSecret:     cfg.JWTSecret,
		JwtExpiration: cfg.JWTExpiration,
	})

	api := e.Group("/api/v1")
	user.NewHandler(db, cfg.JWTSecret).RegisterRoutes(api)

	// graceful shutdown (espera as requisições em andamento serem concluídas antes de sair)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sc := echo.StartConfig{
		Address:         ":" + cfg.Port,
		HideBanner:      false,
		HidePort:        false,
		GracefulTimeout: 5 * time.Second,
	}

	return sc.Start(ctx, e)
}
