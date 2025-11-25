package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	gwhttp "github.com/Xausdorf/pr-reviewer-assignment/internal/gateway/http"
	repopg "github.com/Xausdorf/pr-reviewer-assignment/internal/repository/postgres"
	"github.com/Xausdorf/pr-reviewer-assignment/internal/usecase"
	"github.com/Xausdorf/pr-reviewer-assignment/pkg/migrate"
	pg "github.com/Xausdorf/pr-reviewer-assignment/pkg/postgres"
)

const (
	defaultAddr              = ":8080"
	defaultReadHeaderTimeout = 1 * time.Second
	defaultReadTimeout       = 5 * time.Second
	defaultWriteTimeout      = 5 * time.Second
	defaultShutdownTimeout   = 10 * time.Second

	defaultDBMaxConns          = 10
	defaultDBMinConns          = 1
	defaultDBPingTimeout       = 5 * time.Second
	defaultDBHealthCheckPeriod = 10 * time.Second
)

func main() {
	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		if l, err := log.ParseLevel(lvl); err == nil {
			logger.SetLevel(l)
		}
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	// run migrations before creating pgx pool
	migrationsDir := "./migrations"
	if mdir := os.Getenv("MIGRATIONS_DIR"); mdir != "" {
		migrationsDir = mdir
	}
	if err := migrate.RunMigrations(dbURL, migrationsDir, logger); err != nil {
		logger.WithError(err).Fatal("migrations failed")
	}

	pgCfg := pg.Config{
		ConnString:        dbURL,
		MaxConns:          defaultDBMaxConns,
		MinConns:          defaultDBMinConns,
		HealthCheckPeriod: defaultDBHealthCheckPeriod,
		PingTimeout:       defaultDBPingTimeout,
	}
	pool, err := pg.NewPool(ctx, pgCfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("failed to connect to db")
	}
	defer pg.ClosePool(pool)

	// repositories
	prRepo := repopg.NewPRRepository(pool)
	teamRepo := repopg.NewTeamRepository(pool)
	userRepo := repopg.NewUserRepository(pool)

	// services
	prUseCase := usecase.NewPRUseCase(prRepo, userRepo, logger)
	teamUseCase := usecase.NewTeamUseCase(teamRepo, logger)
	userUseCase := usecase.NewUserUseCase(userRepo, logger)

	// http server
	server := gwhttp.NewServer(prUseCase, teamUseCase, userUseCase, logger)
	handler := gwhttp.Handler(server)

	httpServer := &http.Server{
		Addr:              defaultAddr,
		Handler:           handler,
		ReadTimeout:       defaultReadTimeout,
		WriteTimeout:      defaultWriteTimeout,
		ReadHeaderTimeout: defaultReadHeaderTimeout,
	}

	go func() {
		logger.WithField("addr", httpServer.Addr).Info("starting server")
		if err = httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithError(err).Fatal("http server failed")
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down")
	ctxShut, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()
	if err = httpServer.Shutdown(ctxShut); err != nil {
		logger.WithError(err).Error("error during shutdown")
	}
}
