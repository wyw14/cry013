package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wyw14/cry013/internal/application"
	"github.com/wyw14/cry013/internal/config"
	"github.com/wyw14/cry013/internal/platform/demo"
	"github.com/wyw14/cry013/internal/repository/memory"
	pgrepo "github.com/wyw14/cry013/internal/repository/sqlite"
	"github.com/wyw14/cry013/internal/service"
	httptransport "github.com/wyw14/cry013/internal/transport/http"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger, err := zap.NewProduction()
	if cfg.Environment == "development" {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	rootCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var repo application.Repository
	ready := func() error { return nil }
	if cfg.AllowInMemory {
		repo = memory.New()
	} else {
		ctx, stop := context.WithTimeout(rootCtx, 10*time.Second)
		database, openErr := pgrepo.Open(ctx, cfg.DatabaseURL)
		stop()
		if openErr != nil {
			logger.Fatal("connect database", zap.Error(openErr))
		}
		if err := database.Migrate(rootCtx); err != nil {
			logger.Fatal("migrate database", zap.Error(err))
		}
		repo = pgrepo.NewRepository(database.DB)
		ready = func() error {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			return database.Ready(ctx)
		}
	}

	clock := service.RealClock{}
	ids := service.UUIDGenerator{}
	passwords := service.BcryptHasher{}
	tokens := service.JWTManager{Secret: []byte(cfg.JWTSecret), AccessTTL: cfg.AccessTTL}
	if err := demo.Seed(rootCtx, repo, passwords, clock.Now()); err != nil {
		logger.Fatal("seed demo data", zap.Error(err))
	}

	h := &httptransport.Handler{
		Auth:          application.NewAuthService(repo, clock, ids, passwords, tokens, cfg.RefreshTTL),
		Vaults:    application.NewVaultService(repo, clock, ids),
		Entries:         application.NewEntryService(repo, clock, ids),
		Discovery:     application.NewDiscoveryService(repo),
		Collaboration: application.NewCollaborationService(repo, clock, ids),
		Admin:         application.NewAdminService(repo, clock, ids),
		Ready:         ready,
	}
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: httptransport.Router(h, tokens, logger, cfg.RequestTimeout, cfg.CORSOrigins), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		logger.Info("server listening", zap.String("address", cfg.HTTPAddr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("http server", zap.Error(err))
		}
	}()
	<-rootCtx.Done()
	shutdownCtx, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown", zap.Error(err))
		os.Exit(1)
	}
}
