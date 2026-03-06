package main

import (
	"context"
	"log"
	"net/http"
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

	// Graceful shutdown: when ctx is cancelled (Ctrl-C), shutdown the HTTP server.
	go func() {
		<-ctx.Done()
		log.Println("shutting down...")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve: %v", err)
	}
	log.Println("server stopped")
}
