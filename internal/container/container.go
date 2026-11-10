package container

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Lebedeva2006/go-labs/internal/config"
	api "github.com/Lebedeva2006/go-labs/internal/generated"
	"github.com/Lebedeva2006/go-labs/internal/http/middleware"
	"github.com/Lebedeva2006/go-labs/internal/http/response"
	"github.com/Lebedeva2006/go-labs/internal/http/server"
	createh "github.com/Lebedeva2006/go-labs/internal/http/trip/create"
	finishh "github.com/Lebedeva2006/go-labs/internal/http/trip/finish"
	geth "github.com/Lebedeva2006/go-labs/internal/http/trip/get"
	"github.com/Lebedeva2006/go-labs/internal/infra/postgres"
	"github.com/Lebedeva2006/go-labs/internal/infra/postgres/transaction"
	"github.com/Lebedeva2006/go-labs/internal/infra/telemetry"
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
	log := telemetry.NewLogger(cfg.LogLevel)

	shutdown, err := telemetry.Setup(ctx)
	if err != nil {
		return nil, err
	}

	metrics, err := telemetry.NewMetrics()
	if err != nil {
		return nil, err
	}

	pool, err := postgres.NewPool(ctx, cfg)
	if err != nil {
		return nil, err
	}

	tx := transaction.NewTxManager(pool)
	trips := repository.NewTripRepository(pool, cfg.DatabaseQueryTimeout)
	history := repository.NewTripStatusHistoryRepository(pool, cfg.DatabaseQueryTimeout)

	srv := &server.Server{
		Create:       createh.New(createuc.New(trips, history, tx, metrics), log),
		Get:          geth.New(getuc.New(trips), log),
		Finish:       finishh.New(finishuc.New(trips, history, tx, metrics), log),
		Pool:         postgres.PoolPinger{Pool: pool},
		QueryTimeout: cfg.DatabaseQueryTimeout,
		Log:          log,
	}

	router := chi.NewRouter()
	mw, err := middleware.Metrics()
	if err != nil {
		return nil, err
	}
	router.Use(
		middleware.Recover(log),
		middleware.Tracing,
		middleware.Logging(log),
		mw,
	)

	handler := api.HandlerWithOptions(srv, api.ChiServerOptions{
		BaseRouter: router,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			log.WarnContext(r.Context(), "invalid request parameter", "error", err)
			response.BadRequest(w, r, "invalid request parameter")
		},
	})

	return &Container{
		Handler: handler,
		Log:     log,
		Close: func() {
			pool.Close()
			_ = shutdown(context.Background())
		},
	}, nil
}
