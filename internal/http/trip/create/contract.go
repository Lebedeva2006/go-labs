//go:generate mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test
package create

import (
	"context"

	domain "github.com/Lebedeva2006/go-labs/internal/domain/trip"
	"github.com/Lebedeva2006/go-labs/internal/model"
)

type usecase interface {
	Execute(ctx context.Context, in model.CreateTripInput) (domain.Trip, error)
}