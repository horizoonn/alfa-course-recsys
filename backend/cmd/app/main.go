package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/app"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("application failed", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() { <-ctx.Done(); stop() }()
	application, err := app.New(ctx, cfg, logger)
	if err != nil {
		if ctx.Err() != nil && errors.Is(err, context.Canceled) {
			return nil
		}
		return err
	}
	return application.Run(ctx)
}
