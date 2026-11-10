package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lebedeva2006/go-labs/internal/config"
	"github.com/Lebedeva2006/go-labs/internal/container"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	c, err := container.New(context.Background(), cfg)
	if err != nil {
		log.Fatalf("build container: %v", err)
	}
	defer c.Close()

	srv := &http.Server{
		Addr: cfg.HTTPAddr, Handler: c.Handler,
		ReadTimeout: 5 * time.Second, ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errs := make(chan error, 1)
	go func() {
		c.Log.Info("starting server", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()

	select {
	case err := <-errs:
		log.Fatalf("server failed: %v", err)
	case <-ctx.Done():
		sctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(sctx); err != nil {
			c.Log.Error("graceful shutdown failed", "error", err)
			_ = srv.Close()
		}
	}
}