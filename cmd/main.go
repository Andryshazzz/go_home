package main

import (
	"flag"
	"go_home/config"
	"go_home/internal/app"
	"log"
)

// @title           go_home API
// @version         1.0
// @description     User service API.
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	configPath := flag.String("config", "./config/main.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalf("config error: %s", err)
	}

	app.Run(cfg)
}
