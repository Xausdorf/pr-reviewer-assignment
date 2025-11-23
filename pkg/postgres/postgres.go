package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	ConnString string
	MaxConns   int
	MinConns   int
}

func NewPool(ctx context.Context, cfg Config, logger *log.Logger) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(cfg.ConnString)
	if err != nil {
		return nil, err
	}
	if cfg.MaxConns > 0 {
		config.MaxConns = int32(cfg.MaxConns)
	}
	if cfg.MinConns > 0 {
		config.MinConns = int32(cfg.MinConns)
	}

	config.HealthCheckPeriod = 5 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return nil, err
	}

	logger.WithField("component", "postgres").Info("connected to postgres")
	return pool, nil
}

func ClosePool(p *pgxpool.Pool) {
	if p != nil {
		p.Close()
	}
}
