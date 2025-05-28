package main

import (
	"context"
	"os"
	"os/signal"
	"revnoa/agent"
	"revnoa/config"
	"revnoa/handlers"
	"revnoa/utils"
	"syscall"
)

func main() {
	utils.InitLogger(false)

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		utils.ErrorLogger.Printf("Invalid config: %v", err)
		os.Exit(1)
	}

	handlers.SetAuthKey(cfg.API.AuthKey)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	agent.RunAgent(ctx, cfg, cfg.UUID)
}
