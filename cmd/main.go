package main

import (
	"flag"
	"go_home/config"
	"go_home/internal/app"
	"log"
)

func main() {
	configPath := flag.String("config", "./config/main.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalf("config error: %s", err)
	}

	app.Run(cfg)
}
