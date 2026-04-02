FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /auth-service ./cmd/auth-service

FROM debian:bookworm-slim

RUN useradd --system --uid 10001 --no-create-home --shell /usr/sbin/nologin app

WORKDIR /app

COPY --from=builder /auth-service /usr/local/bin/auth-service

USER app

EXPOSE 8000

ENTRYPOINT ["auth-service"]
