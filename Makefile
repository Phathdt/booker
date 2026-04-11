# Source .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: build run-users-svc migrate-up migrate-down migrate-status seed \
       sqlc-generate proto-generate proto-lint swagger \
       docker-build docker-up docker-down docker-logs \
       test lint format format-check mock hooks

# Build
build:
	@mkdir -p bin
	go build -o bin/main .

# Run services
run-users-svc:
	go run . users-svc --config config.yml

# Database
migrate-up:
	go run . migrate --config config.yml up

migrate-down:
	go run . migrate --config config.yml down

migrate-status:
	go run . migrate --config config.yml status

seed:
	psql -h localhost -U postgres -d booker -f seed.sql

# Code generation
sqlc-generate:
	sqlc generate

proto-generate:
	buf dep update && buf generate

proto-lint:
	buf lint

swagger:
	swag init --parseDependency --parseInternal -g main.go

# Docker
docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

# Format
format:
	@echo "Formatting Go files..."
	@gofmt -w .
	@echo "Organizing imports..."
	@goimports -w .
	@echo "Formatting line lengths..."
	@golines -w -m 120 .
	@echo "Applying gofumpt..."
	@gofumpt -extra -w .
	@echo "Formatting YAML files..."
	@npx prettier --write "*.yml" "*.yaml" "docker-compose.yml" 2>/dev/null || true
	@echo "Done!"

format-check:
	@gofmt -l . | grep -q . && echo "Run 'make format' to fix" && exit 1 || echo "Go formatted OK"
	@npx prettier --check "*.yml" "*.yaml" "docker-compose.yml" 2>/dev/null || (echo "Run 'make format' to fix YAML" && exit 1)

# Testing
test:
	go test ./... -v

test-unit:
	go test ./modules/... -v -short

test-integration:
	go test ./test/... -v -timeout 300s

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Mocks
mock:
	mockery

# Git hooks (run once to install)
hooks:
	@echo "Installing git pre-commit hook..."
	@mkdir -p .git/hooks
	@cp scripts/pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Done!"

# Quality
lint:
	golangci-lint run
