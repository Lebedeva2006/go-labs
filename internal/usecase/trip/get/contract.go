//go:generate mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test
package get

import (
	"context"

	domain "github.com/Lebedeva2006/go-labs/internal/domain/trip"
)

type tripRepository interface {
	GetByID(ctx context.Context, id string) (domain.Trip, error)
}