APP_NAME=auth-service

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
