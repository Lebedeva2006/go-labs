//go:generate mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test
package create

import (
	"context"

	domain "github.com/Lebedeva2006/go-labs/internal/domain/trip"
)

type (
	tripRepository interface {
		Create(ctx context.Context, trip domain.Trip) error
	}
	historyRepository interface {
		Create(ctx context.Context, tripID string, from *domain.TripStatus, to domain.TripStatus, reason string) error
	}
	txManager interface {
		Do(ctx context.Context, fn func(ctx context.Context) error) error
	}
)
