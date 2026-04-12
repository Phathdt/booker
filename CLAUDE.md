# CLAUDE.md

## Project Overview

**Booker** — CEX (Centralized Exchange) demo for token trading, similar to Binance.

### Architecture

- **1 binary, multi-command**: `urfave/cli` — each service is a CLI command (`./main users-svc`, `./main wallet-svc`, etc.)
- **External**: Traefik → Fiber REST (validate input, respond with traceId)
- **Internal**: Service-to-service via gRPC
- **Observability**: OpenTelemetry → OTel Collector → Grafana Tempo (traces) + Loki (logs)

### Tech Stack

Go 1.26, Fiber (REST), gRPC (inter-service), SQLC (Postgres), Redis, NATS JetStream, Traefik v3, buf (protobuf), goose (migrations), mockery (mocks), testcontainers (integration tests), go-playground/validator (input validation)

### Services

| Service          | Command                   | HTTP  | gRPC   | Status |
| ---------------- | ------------------------- | ----- | ------ | ------ |
| users-svc        | `./main users-svc`        | :8081 | :50051 | Done   |
| wallet-svc       | `./main wallet-svc`       | :8082 | :50052 | Done   |
| order-svc        | `./main order-svc`        | :8083 | :50053 | Done   |
| matching-svc     | `./main matching-svc`     | :8084 | :50054 | Done   |
| market-svc       | `./main market-svc`       | :8085 | —      | Done   |
| notification-svc | `./main notification-svc` | :8086 | —      | Done   |

## Code Structure

```
main.go              # CLI commands (urfave/cli)
cli/                 # Service runners (wiring + lifecycle)
cmd/
  grpc/{name}/       # gRPC server implementations (inter-service only)
  http/{name}/       # Fiber REST handlers (1 file per endpoint, closure factory pattern)
  shared/            # Shared providers (DB, Redis, OTel, logger)
modules/{name}/
  domain/entities/   # Domain structs
  domain/interfaces/ # Repository + service interfaces
  domain/errors.go   # Domain-specific errors
  application/dto/   # Request DTOs (with validator tags, used by Fiber handlers)
  application/services/  # Business logic
  application/usecases/  # Use case orchestration
  infrastructure/gen/    # SQLC generated code
  infrastructure/query/  # SQL query files
  infrastructure/repositories/ # Repository implementations
  infrastructure/token/  # JWT token service (users only)
pkg/
  errors/            # AppError interface + constructors
  httpserver/         # Fiber middleware (tracing, requestId, auth, error handler, response, BindAndValidate)
  interceptors/      # gRPC interceptors (logging, user header)
  logger/            # slog-based structured logger
  otel/              # OpenTelemetry SDK setup
proto/{name}/v1/     # Protobuf definitions (inter-service only)
config/              # Viper config loader
migrations/          # Goose SQL migrations
infra/               # Docker infra configs (otel, tempo, loki, grafana)
test/testcontainers/ # Postgres + Redis test containers
```

## Commands

```bash
make build              # Build binary to bin/main
make run-users-svc      # Run users service locally
make migrate-up         # Run migrations
make migrate-down       # Rollback last migration
make seed               # Seed assets, trading pairs + 100 test accounts
make sqlc-generate      # Regenerate SQLC code
make proto-generate     # Regenerate protobuf
make mock               # Regenerate mockery mocks
make test               # Run all tests
make test-unit          # Run unit tests only (fast)
make test-integration   # Run integration tests (needs Docker)
make test-coverage      # Generate coverage report
make format             # Format Go + YAML files
make docker-build       # Build Docker image
make docker-up          # Start all services
make docker-down        # Stop all services
```

## Conventions

- **Clean architecture**: domain → application → infrastructure
- **Constructor DI**: No framework, interface-based
- **HTTP handlers**: Closure factory pattern — `func Register(uc) fiber.Handler` — 1 file per endpoint
- **Validation**: `go-playground/validator` tags on DTOs, `httpserver.BindAndValidate()` in handlers
- **Response format**: `{ data, error, trace_id, request_id }`
- **Proto**: Inter-service gRPC only, no REST annotations
- **Errors**: `pkg/errors.AppError` → mapped to HTTP status in Fiber error handler, gRPC codes in gRPC handler
- **Testing**: Mockery mocks for unit tests, testcontainers for integration tests
- **Decimal**: `shopspring/decimal` for financial values, `NUMERIC` in Postgres
