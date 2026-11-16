package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/client/textsearch"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/config"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/curation"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
	catalogrepository "github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/catalog"
	platformhistoryrepository "github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/platformhistory"
	catalogservice "github.com/horizoonn/alfa-course-recsys/backend/internal/service/catalog"
	platformhistoryservice "github.com/horizoonn/alfa-course-recsys/backend/internal/service/platformhistory"
	"github.com/horizoonn/alfa-course-recsys/backend/internal/service/recommendation"
	httpapi "github.com/horizoonn/alfa-course-recsys/backend/internal/transport/http"
	"github.com/horizoonn/alfa-course-recsys/backend/platform/postgres"
)

type App struct {
	server          *http.Server
	logger          *slog.Logger
	shutdownTimeout time.Duration
	ready           atomic.Bool
	address         atomic.Value
}

func New(ctx context.Context, cfg config.Config, logger *slog.Logger) (*App, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	startupCtx, cancel := context.WithTimeout(ctx, cfg.StartupTimeout)
	defer cancel()
	pool, err := postgres.NewPool(startupCtx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()
	txManager := postgres.NewTxManager(pool)
	repository := catalogrepository.NewRepository(txManager.Executor)
	loader := catalogservice.NewService(repository, txManager)
	materials, err := loader.Load(startupCtx)
	if err != nil {
		return nil, fmt.Errorf("initialize catalog: %w", err)
	}
	historyProfiles, err := platformhistoryservice.NewService(platformhistoryrepository.NewRepository(pool)).Load(startupCtx, materials)
	if err != nil {
		return nil, fmt.Errorf("load platform history: %w", err)
	}
	return appFromMaterials(startupCtx, cfg, logger, materials, historyProfiles)
}

func appFromMaterials(
	ctx context.Context,
	cfg config.Config,
	logger *slog.Logger,
	materials []domain.Material,
	historyProfiles []domain.HistoryProfile,
) (*App, error) {
	started := time.Now()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if materials == nil {
		materials = []domain.Material{}
	}
	curated, err := curation.Load(cfg.CurationPath)
	if err != nil {
		return nil, err
	}
	materials, evidence, err := catalogservice.PrepareFeatures(materials, curated.Reviewed)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	application := &App{logger: logger, shutdownTimeout: cfg.ShutdownTimeout}
	starter, histories := curated.Active(evidence)
	corpus := recommendation.NewSearchCorpus(materials)
	scores, err := recommendation.LoadTextScores(cfg.TextScoresPath, corpus)
	if err != nil {
		return nil, fmt.Errorf("initialize text scores: %w", err)
	}
	logger.Info("text scores loaded", "documents", len(corpus.Documents), "tasks", len(scores), "duration", time.Since(started))
	search, err := textsearch.New(cfg.SearchURL, cfg.SearchMode, corpus.ID, cfg.SearchTimeout)
	if err != nil {
		return nil, err
	}
	recommender, err := recommendation.NewService(materials, starter, histories, scores, search, historyProfiles)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	handler, err := httpapi.NewServer(func() bool { return application.ready.Load() }, recommender, logger)
	if err != nil {
		return nil, fmt.Errorf("initialize HTTP: %w", err)
	}
	application.server = &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	withTasks := 0
	for _, material := range materials {
		if len(material.Features.TaskFits) > 0 {
			withTasks++
		}
	}
	logger.Info("catalog loaded", "materials", len(materials), "with_task_fits", withTasks, "duration", time.Since(started))
	return application, nil
}

func (a *App) Run(ctx context.Context) error {
	if ctx.Err() != nil {
		return nil
	}
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", a.server.Addr)
	if err != nil {
		return fmt.Errorf("listen HTTP: %w", err)
	}
	defer listener.Close()
	a.address.Store(listener.Addr().String())
	a.ready.Store(true)
	defer a.ready.Store(false)
	serveErr := make(chan error, 1)
	go func() { serveErr <- a.server.Serve(listener) }()
	a.logger.Info("HTTP server started", "address", listener.Addr().String())
	var runErr error
	returned := false
	select {
	case runErr = <-serveErr:
		returned = true
	case <-ctx.Done():
	}
	a.ready.Store(false)
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), a.shutdownTimeout)
	defer cancel()
	shutdownErr := a.server.Shutdown(shutdownCtx)
	if shutdownErr != nil {
		shutdownErr = errors.Join(fmt.Errorf("shutdown HTTP: %w", shutdownErr), a.server.Close())
	}
	if !returned {
		runErr = <-serveErr
	}
	if errors.Is(runErr, http.ErrServerClosed) {
		runErr = nil
	}
	a.logger.Info("HTTP server stopped")
	return errors.Join(runErr, shutdownErr)
}
