package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/darthxd/agendac/internal/config"
	"github.com/darthxd/agendac/internal/database"
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

	e := echo.New()
	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// graceful shutdown (espera as requisições em andamento serem concluídas antes de sair)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sc := echo.StartConfig{
		Address:         ":" + cfg.Port,
		GracefulTimeout: 5 * time.Second,
	}

	return sc.Start(ctx, e)
}
