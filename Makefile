APP_NAME=auth-service
GOOSE_DRIVER=postgres

ifneq (,$(wildcard .env))
include .env
export
endif

PG_HOST ?= localhost
PG_PORT ?= 5432
PG_USER ?= postgres
PG_PASSWORD ?= postgres
PG_DB_NAME ?= auth_service
PG_SSLMODE ?= disable

GOOSE_DBSTRING=host=$(PG_HOST) port=$(PG_PORT) user=$(PG_USER) password=$(PG_PASSWORD) dbname=$(PG_DB_NAME) sslmode=$(PG_SSLMODE)

run:
	go run ./cmd/$(APP_NAME)

build:
	mkdir -p ./bin
	go build -o ./bin/$(APP_NAME) ./cmd/$(APP_NAME)

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

migrate-create:
	@test -n "$(NAME)" && goose -dir ./migrations create $(NAME) sql

migrate-up:
	goose -dir ./migrations $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" up

migrate-down:
	goose -dir ./migrations $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" down

migrate-status:
	goose -dir ./migrations $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" status
