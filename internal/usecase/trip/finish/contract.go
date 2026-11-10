//go:generate mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test
package finish

import (
	"context"
	"time"

	domain "github.com/Lebedeva2006/go-labs/internal/domain/trip"
)

type (
	tripRepository interface {
		GetByID(ctx context.Context, id string) (domain.Trip, error)
		Finish(ctx context.Context, id string, finishedAt time.Time) error
	}
	historyRepository interface {
		Create(ctx context.Context, tripID string, from *domain.TripStatus, to domain.TripStatus, reason string) error
	}
	txManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) error
	}
)
