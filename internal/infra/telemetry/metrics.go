package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type Metrics struct {
	created   metric.Int64Counter
	completed metric.Int64Counter
}

func NewMetrics() (*Metrics, error) {
	m := otel.Meter("trip-service")
	c, err := m.Int64Counter("tripgo.trips.created", metric.WithDescription("Number of created trips"))
	if err != nil {
		return nil, err
	}
	d, err := m.Int64Counter("tripgo.trips.completed", metric.WithDescription("Number of completed trips"))
	if err != nil {
		return nil, err
	}
	return &Metrics{created: c, completed: d}, nil
}

func (m *Metrics) TripCreated(ctx context.Context) {
	m.created.Add(ctx, 1)
}

func (m *Metrics) TripCompleted(ctx context.Context) {
	m.completed.Add(ctx, 1)
}
