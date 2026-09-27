package main

import (
 "log"
 "net/http"
 "time"

 "github.com/go-chi/chi/v5"

 "github.com/Lebedeva2006/go-labs/internal/config"
)

func main() {
 cfg, err := config.Load()
 if err != nil {
  log.Fatalf("failed to load config: %v", err)
 }

 router := chi.NewRouter()

 router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "application/json")
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

 log.Printf("starting server on %s", cfg.HTTPAddr)
 if err := server.ListenAndServe(); err != nil {
  log.Fatalf("server failed: %v", err)
 }
}
