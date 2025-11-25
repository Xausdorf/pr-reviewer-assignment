# Build
FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/app ./cmd/server

# Run
FROM alpine:3.18
WORKDIR /app

COPY --from=builder /out/app /app/app
COPY --from=builder /src/migrations /app/migrations

EXPOSE 8080
ENTRYPOINT ["/app/app"]