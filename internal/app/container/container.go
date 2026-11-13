package container

import (
	"context"
	"go_home/config"
)

type Container struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Container {
	return &Container{cfg: cfg}
}

func (c *Container) Start(ctx context.Context) error {
	return nil
}

func (c *Container) Stop() {}
