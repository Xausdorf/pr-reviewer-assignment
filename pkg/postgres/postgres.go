package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

const defaultPingTimeout = 5 * time.Second

type Config struct {
	ConnString        string
	MaxConns          int32
	MinConns          int32
	HealthCheckPeriod time.Duration
	PingTimeout       time.Duration
}

func NewPool(ctx context.Context, cfg Config, logger *log.Logger) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(cfg.ConnString)
	if err != nil {
		logger.WithError(err).Error("Failed to parse postgres config")
		return nil, err
	}
	if cfg.MaxConns > 0 {
		config.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns > 0 {
		config.MinConns = cfg.MinConns
	}
	if cfg.HealthCheckPeriod != 0 {
		config.HealthCheckPeriod = cfg.HealthCheckPeriod
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		logger.WithError(err).Error("Failed to create postgres connection pool")
		return nil, err
	}

	pingTimeout := defaultPingTimeout
	if cfg.PingTimeout != 0 {
		pingTimeout = cfg.PingTimeout
	}
	ctxPing, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	logger.WithField("component", "postgres").Info("Pinging postgres")
	if err := pool.Ping(ctxPing); err != nil {
		logger.WithError(err).Error("Failed to ping postgres")
		pool.Close()
		return nil, err
	}

	logger.WithField("component", "postgres").Info("Connected to postgres")
	return pool, nil
}

func ClosePool(p *pgxpool.Pool) {
	if p != nil {
		p.Close()
	}
}
