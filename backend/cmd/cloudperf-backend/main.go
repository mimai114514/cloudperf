package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"cloudperf/backend/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := app.LoadConfig()
	srv, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("init backend: %v", err)
	}
	defer srv.Close()

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
