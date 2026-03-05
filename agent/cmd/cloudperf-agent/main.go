package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"cloudperf/agent/internal/client"
	"cloudperf/agent/internal/config"
)

func main() {
	cfg := config.Load()
	if !cfg.Valid() {
		log.Fatal("invalid config: BACKEND_WS_URL, NODE_ID, NODE_TOKEN are required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a := client.New(cfg)
	a.Run(ctx)
}
