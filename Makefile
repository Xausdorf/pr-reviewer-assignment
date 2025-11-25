include .env
export

.PHONY: build
build: ## Build binary for linux (output `bin/app`)
	GOOS=linux GOARCH=amd64 go build -v -o bin/app ./cmd/server

.PHONY: docker-build
docker-build: ## Build docker image locally
	docker build -t pr-reviewer:latest .

.PHONY: up
up: ### Run docker-compose
	docker-compose up --build -d && docker-compose logs -f

.PHONY: down
down: ### Down docker-compose
	docker-compose down --remove-orphans

.PHONY: docker-rm-volume
docker-rm-volume: ### remove docker volume
	docker volume rm pg-data || true

.PHONY: run
run: ## Run locally (requires env variables)
	go run ./cmd/server

.PHONY: fmt
fmt: ## go fmt
	go fmt ./...

.PHONY: linter-golangci
linter-golangci: ### check by golangci linter
	golangci-lint run

.PHONY: test
test: ### run tests
	go test -v ./...

.PHONY: cover-html
cover-html: ### run test with coverage and open html report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out

.PHONY: cover
cover: ### run test with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	rm coverage.out

.PHONY: oapi-gen
oapi-gen: ### generate types and chi server from OpenAPI (oapi-codegen)
	oapi-codegen -generate types -o internal/gateway/http/openapi_types.gen.go -package http api/openapi.yml
	oapi-codegen -generate chi-server -o internal/gateway/http/openapi_server.gen.go -package http api/openapi.yml
