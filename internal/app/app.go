package app

import (
	"context"
	"go_home/config"
	"go_home/internal/app/container"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func Run(cfg *config.Config) {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM, syscall.SIGINT,
	)
	defer cancel()

	cnt := container.New(cfg)

	if err := cnt.Start(ctx); err != nil {
		log.Printf("start error: %v", err)
		cnt.Stop()
		os.Exit(1)
	}

	<-ctx.Done()
	log.Println("shutting down...")

	cnt.Stop()
	log.Println("application stopped")
}
