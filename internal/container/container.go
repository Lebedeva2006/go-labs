package container

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/Lebedeva2006/go-labs/internal/config"
	api "github.com/Lebedeva2006/go-labs/internal/generated"
	createh "github.com/Lebedeva2006/go-labs/internal/http/trip/create"
	finishh "github.com/Lebedeva2006/go-labs/internal/http/trip/finish"
	geth "github.com/Lebedeva2006/go-labs/internal/http/trip/get"
	"github.com/Lebedeva2006/go-labs/internal/http/response"
	"github.com/Lebedeva2006/go-labs/internal/http/server"
	"github.com/Lebedeva2006/go-labs/internal/infra/postgres"
	"github.com/Lebedeva2006/go-labs/internal/infra/postgres/transaction"
	repository "github.com/Lebedeva2006/go-labs/internal/repository/trip"
	createuc "github.com/Lebedeva2006/go-labs/internal/usecase/trip/create"
	finishuc "github.com/Lebedeva2006/go-labs/internal/usecase/trip/finish"
	getuc "github.com/Lebedeva2006/go-labs/internal/usecase/trip/get"
)

type Container struct {
	Handler http.Handler
	Log     *slog.Logger
	Close   func()
}

func New(ctx context.Context, cfg config.Config) (*Container, error) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parseLevel(cfg.LogLevel)}))

	pool, err := postgres.NewPool(ctx, cfg)
	if err != nil {
		return nil, err
	}
	tx := transaction.NewTxManager(pool)
	trips := repository.NewTripRepository(pool, cfg.DatabaseQueryTimeout)
	history := repository.NewTripStatusHistoryRepository(pool, cfg.DatabaseQueryTimeout)

	srv := &server.Server{
		Create:       createh.New(createuc.New(trips, history, tx), log),
		Get:          geth.New(getuc.New(trips), log),
		Finish:       finishh.New(finishuc.New(trips, history, tx), log),
		Pool:         postgres.PoolPinger{Pool: pool},
		QueryTimeout: cfg.DatabaseQueryTimeout,
		Log:          log,
	}

	handler := api.HandlerWithOptions(srv, api.ChiServerOptions{
		BaseRouter: chi.NewRouter(),
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			log.WarnContext(r.Context(), "invalid request parameter", "error", err)
			response.BadRequest(w, r, "invalid request parameter")
		},
	})
	return &Container{Handler: handler, Log: log, Close: pool.Close}, nil
}

func parseLevel(s string) slog.Level {
	var l slog.Level
	if err := l.UnmarshalText([]byte(s)); err != nil {
		return slog.LevelInfo
	}
	return l
}