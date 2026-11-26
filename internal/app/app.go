package app

import (
	"context"
	"fmt"
	"go_home/config"
	"go_home/internal/app/container"
	"os"
	"os/signal"
	"syscall"
)

func Run(cfg *config.Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	c := container.New(cfg)
	defer c.Stop()

	if err := c.Start(ctx); err != nil {
		return fmt.Errorf("run app: %w", err)
	}

	return nil
}
