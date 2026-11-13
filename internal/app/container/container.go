package container

import (
	"context"
	"fmt"
	"go_home/config"
	"go_home/pkg/database"
	"go_home/pkg/logger"
)

type Container struct {
	cfg    *config.Config
	logger *logger.Logger
	pool   *database.ConnectionPool
}

func New(cfg *config.Config) *Container {
	return &Container{cfg: cfg}
}

func (c *Container) Start(ctx context.Context) error {
	if err := c.initLogger(); err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	if err := c.initDatabase(ctx); err != nil {
		return fmt.Errorf("init database: %w", err)
	}

	return nil
}

func (c *Container) Stop() {
	if c.pool != nil {
		c.pool.Close()
	}

	if c.logger != nil {
		c.logger.Close()
	}
}

func (c *Container) initLogger() error {
	log, err := logger.NewLogger(logger.Config{
		Level:  c.cfg.Logger.Level,
		Folder: c.cfg.Logger.Folder,
	})
	if err != nil {
		return err
	}

	c.logger = log

	return nil
}

func (c *Container) initDatabase(ctx context.Context) error {
	pool, err := database.NewConnectionPool(ctx, database.Config{
		Host:     c.cfg.Postgres.Host,
		Port:     c.cfg.Postgres.Port,
		User:     c.cfg.Postgres.User,
		Password: c.cfg.Postgres.Password,
		DB:       c.cfg.Postgres.DB,
		SSLMode:  c.cfg.Postgres.SSLMode,
		Timeout:  c.cfg.Postgres.Timeout,
	})
	if err != nil {
		return err
	}
	c.pool = pool
	return nil
}
