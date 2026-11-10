//go:generate mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test
package finish

import (
	"context"
	"github.com/google/uuid"
	domain "github.com/Lebedeva2006/go-labs/internal/domain/trip"
)

type usecase interface {
	Execute(ctx context.Context, id uuid.UUID) (domain.Trip, error)
}