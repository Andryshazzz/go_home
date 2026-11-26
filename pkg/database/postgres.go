package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
	SSLMode  string
	Timeout  time.Duration
}

type ConnectionPool struct {
	*pgxpool.Pool
	opTimeout time.Duration
}

func (c Config) dsn() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     net.JoinHostPort(c.Host, c.Port),
		Path:     c.DB,
		RawQuery: url.Values{"sslmode": {c.SSLMode}}.Encode(),
	}

	return u.String()
}

func NewConnectionPool(
	ctx context.Context,
	cfg Config,
) (*ConnectionPool, error) {
	pgxconfig, err := pgxpool.ParseConfig(cfg.dsn())

	if err != nil {
		return nil, fmt.Errorf("parse pgxconfig: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxconfig)

	if err != nil {
		return nil, fmt.Errorf("create pgxpool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()

		return nil, fmt.Errorf("pgxpool ping: %w", err)
	}

	return &ConnectionPool{
		Pool:      pool,
		opTimeout: cfg.Timeout,
	}, nil
}

func (p *ConnectionPool) OpTimeout() time.Duration {
	return p.opTimeout
}
