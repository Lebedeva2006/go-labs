package config

import (
 "fmt"
 "os"
 "strconv"
 "time"
)


type Config struct {
 HTTPAddr        string
 DatabaseURL     string
 LogLevel        string
 ShutdownTimeout time.Duration

 DatabaseMaxConns        int
 DatabaseMinConns        int
 DatabaseConnectTimeout  time.Duration
 DatabaseQueryTimeout    time.Duration
 DatabaseMaxConnLifetime time.Duration
}


func Load() (Config, error) {
 var cfg Config
 var err error

 cfg.HTTPAddr, err = requireString("HTTP_ADDR")
 if err != nil {
  return Config{}, err
 }

 cfg.DatabaseURL, err = requireString("DATABASE_URL")
 if err != nil {
  return Config{}, err
 }

 cfg.LogLevel, err = requireString("LOG_LEVEL")
 if err != nil {
  return Config{}, err
 }

 cfg.ShutdownTimeout, err = requireDuration("SHUTDOWN_TIMEOUT")
 if err != nil {
  return Config{}, err
 }

 cfg.DatabaseMaxConns, err = requireInt("DATABASE_MAX_CONNS")
 if err != nil {
  return Config{}, err
 }

 cfg.DatabaseMinConns, err = requireInt("DATABASE_MIN_CONNS")
 if err != nil {
  return Config{}, err
 }

 cfg.DatabaseConnectTimeout, err = requireDuration("DATABASE_CONNECT_TIMEOUT")
 if err != nil {
  return Config{}, err
 }

 cfg.DatabaseQueryTimeout, err = requireDuration("DATABASE_QUERY_TIMEOUT")
 if err != nil {
  return Config{}, err
 }

 cfg.DatabaseMaxConnLifetime, err = requireDuration("DATABASE_MAX_CONN_LIFETIME")
 if err != nil {
  return Config{}, err
 }

 return cfg, nil
}

func requireString(key string) (string, error) {
 value, ok := os.LookupEnv(key)
 if !ok || value == "" {
  return "", fmt.Errorf("config: required env var %s is not set", key)
 }
 return value, nil
}

func requireInt(key string) (int, error) {
 raw, err := requireString(key)
 if err != nil {
  return 0, err
 }
 value, err := strconv.Atoi(raw)
 if err != nil {
  return 0, fmt.Errorf("config: env var %s must be an integer: %w", key, err)
 }
 return value, nil
}

func requireDuration(key string) (time.Duration, error) {
 raw, err := requireString(key)
 if err != nil {
  return 0, err
 }
 value, err := time.ParseDuration(raw)
 if err != nil {
  return 0, fmt.Errorf("config: env var %s must be a valid duration (e.g. \"5s\"): %w", key, err)
 }
 return value, nil
}
