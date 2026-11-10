.PHONY: migrate migrate-down run generate build test test-integration lint fmt

migrate:
	go tool goose -dir migrations postgres "$$DATABASE_URL" up

migrate-down:
	go tool goose -dir migrations postgres "$$DATABASE_URL" down

run:
	go run ./cmd/service

generate:
	go tool oapi-codegen \
	 -generate types,chi-server \
	 -include-operation-ids createTrip,getTrip,finishTrip,health,ready \
	 -package api \
	 -o internal/generated/api.gen.go \
 	 contracts/openapi/trip-service.openapi.yaml

build:
	go build -o bin/service ./cmd/service

lint:
	go tool golangci-lint run ./...

fmt:
	go tool golangci-lint fmt ./...

test:
	go test -race -cover ./...

test-integration:
	go test -race -tags=integration ./...
