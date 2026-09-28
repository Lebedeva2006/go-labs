package main

import (
 "context"
 "log"
 "net/http"
 "os/signal"
 "syscall"
 "time"

 "github.com/go-chi/chi/v5"

 "github.com/Lebedeva2006/go-labs/internal/config"
 "github.com/Lebedeva2006/go-labs/internal/db"
 api "github.com/Lebedeva2006/go-labs/internal/generated"
 "github.com/Lebedeva2006/go-labs/internal/handler"
 "github.com/Lebedeva2006/go-labs/internal/repository"
)

func main() {
 cfg, err := config.Load()
 if err != nil {
  log.Fatalf("failed to load config: %v", err)
 }

 startupCtx := context.Background()

 pool, err := db.NewPool(startupCtx, cfg)
 if err != nil {
  log.Fatalf("failed to connect to database: %v", err)
 }
 defer pool.Close()
 txManager := db.NewTxManager(pool)
 tripRepo := repository.NewTripRepository(pool)
 historyRepo := repository.NewTripStatusHistoryRepository(pool)

 server := &handler.Server{
  Pool:         db.PoolPinger{Pool: pool},
  TxManager:    txManager,
  TripRepo:     tripRepo,
  HistoryRepo:  historyRepo,
  QueryTimeout: cfg.DatabaseQueryTimeout,
 }

 router := chi.NewRouter()
 apiHandler := api.HandlerFromMux(server, router)

 httpServer := &http.Server{
  Addr:              cfg.HTTPAddr,
  Handler:           apiHandler,
  ReadTimeout:       5 * time.Second,
  ReadHeaderTimeout: 5 * time.Second,
  WriteTimeout:      10 * time.Second,
  IdleTimeout:       60 * time.Second,
 }

 ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
 defer stop()

 serverErrors := make(chan error, 1)
 go func() {
  log.Printf("starting server on %s", cfg.HTTPAddr)
  if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
   serverErrors <- err
  }
 }()

 select {
 case err := <-serverErrors:
  log.Fatalf("server failed: %v", err)
 case <-ctx.Done():
  log.Println("shutdown signal received")

  shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
  defer cancel()

  if err := httpServer.Shutdown(shutdownCtx); err != nil {
   log.Printf("graceful shutdown failed: %v", err)
   if closeErr := httpServer.Close(); closeErr != nil {
    log.Printf("forced close failed: %v", closeErr)
   }
  } else {
   log.Println("server stopped gracefully")
  }
 }
}
