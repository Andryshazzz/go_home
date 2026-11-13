package container

import (
	"context"
	"fmt"
	"go_home/config"
	"go_home/pkg/logger"
)

type Container struct {
	cfg    *config.Config
	logger *logger.Logger
}

func New(cfg *config.Config) *Container {
	return &Container{cfg: cfg}
}

func (c *Container) Start(ctx context.Context) error {
	if err := c.initLogger(); err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	return nil
}

func (c *Container) Stop() {
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
