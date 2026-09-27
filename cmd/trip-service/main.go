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

 router := chi.NewRouter()

 router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  w.Write([]byte(`{"status":"ok"}`))
 })

 router.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")

  if err := db.Ping(r.Context(), pool, cfg.DatabaseQueryTimeout); err != nil {
   log.Printf("readiness check failed: %v", err)
   w.WriteHeader(http.StatusServiceUnavailable)
   w.Write([]byte(`{"status":"unavailable"}`))
   return
  }

  w.WriteHeader(http.StatusOK)
  w.Write([]byte(`{"status":"ok"}`))
 })


 server := &http.Server{
  Addr:              cfg.HTTPAddr,
  Handler:           router,
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
  if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

  if err := server.Shutdown(shutdownCtx); err != nil {
   log.Printf("graceful shutdown failed: %v", err)
   if closeErr := server.Close(); closeErr != nil {
    log.Printf("forced close failed: %v", closeErr)
   }
  } else {
   log.Println("server stopped gracefully")
  }
 }
}
