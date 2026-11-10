package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/Lebedeva2006/go-labs/internal/http/response"
)

func route(r *http.Request) string {
	if rc := chi.RouteContext(r.Context()); rc != nil {
		if p := rc.RoutePattern(); p != "" {
			return p
		}
	}
	return "unmatched"
}

func Recover(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if p := recover(); p != nil {
					log.ErrorContext(r.Context(), "panic", "panic", p, "stack", string(debug.Stack()))
					response.Error(w, r, log, errPanic)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func Tracing(next http.Handler) http.Handler {
	tracer := otel.Tracer("trip-service/http")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span := tracer.Start(ctx, r.Method, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()
		sw := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r.WithContext(ctx))
		rt := route(r)
		span.SetName(r.Method + " " + rt)
		span.SetAttributes(attribute.String("http.request.method", r.Method), attribute.String("http.route", rt),
			attribute.Int("http.response.status_code", sw.status))
		if sw.status >= 500 {
			span.SetStatus(codes.Error, "server error")
			span.SetAttributes(attribute.String("error.type", http.StatusText(sw.status)))
		}
	})
}

func Logging(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w}
			next.ServeHTTP(sw, r)
			lvl := slog.LevelInfo
			if sw.status >= 500 {
				lvl = slog.LevelError
			}
			log.Log(r.Context(), lvl, "request handled",
				"method", r.Method, "route", route(r), "status", sw.status,
				"duration_ms", time.Since(start).Milliseconds())
		})
	}
}

func Metrics() (func(http.Handler) http.Handler, error) {
	m := otel.Meter("trip-service/http")
	dur, err := m.Float64Histogram("http.server.request.duration", metric.WithUnit("s"))
	if err != nil {
		return nil, err
	}
	active, err := m.Int64UpDownCounter("http.server.active_requests", metric.WithUnit("{request}"))
	if err != nil {
		return nil, err
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			base := metric.WithAttributes(attribute.String("http.request.method", r.Method))
			active.Add(r.Context(), 1, base)
			defer active.Add(r.Context(), -1, base)
			sw := &statusWriter{ResponseWriter: w}
			next.ServeHTTP(sw, r)
			dur.Record(r.Context(), time.Since(start).Seconds(), metric.WithAttributes(
				attribute.String("http.request.method", r.Method),
				attribute.String("http.route", route(r)),
				attribute.Int("http.response.status_code", sw.status),
			))
		})
	}, nil
}

var errPanic = errors.New("panic")
