# Booker

CEX (Centralized Exchange) demo — token trading platform similar to Binance.

## Architecture

```
Client → Traefik (:80) → Fiber REST → Usecase → PostgreSQL/Redis
                                    → gRPC (inter-service) → Other services
```

- **External**: Fiber REST with input validation, traceId in response
- **Internal**: gRPC between services
- **Observability**: OpenTelemetry → Tempo (traces) + Loki (logs) → Grafana

## Quick Start

```bash
# Prerequisites: Go 1.26+, Docker, sqlc, buf, mockery

# Start infrastructure
docker compose up -d postgres-db redis-db nats

# Run migrations + seed
make migrate-up
make seed

# Start users service
make run-users-svc
```

## Docker (full stack)

```bash
docker compose up --build -d
```

Services available:

- API: http://localhost (via Traefik)
- Grafana: http://localhost:3000 (admin/admin)
- Traefik Dashboard: http://localhost:8888
- NATS Monitoring: http://localhost:8222

## API Endpoints

### Validation

Input validated at handler level via `go-playground/validator`:
- `email`: required, valid format
- `password`: required, 8-72 characters

### Auth (public)

```
POST /api/v1/auth/register  { email, password }
POST /api/v1/auth/login     { email, password }
POST /api/v1/auth/refresh   { refresh_token }
```

### Auth (protected — Bearer token required)

```
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

### Users (protected)

```
GET /api/v1/users/:id
GET /api/v1/users?limit=20&offset=0
```

### Response format

```json
{
  "data": { ... },
  "trace_id": "abc123",
  "request_id": "uuid"
}
```

## Development

```bash
make format         # Format code
make test           # Run all tests
make test-unit      # Unit tests only
make test-coverage  # Coverage report
make mock           # Regenerate mocks
make sqlc-generate  # Regenerate SQLC
make proto-generate # Regenerate protobuf
```

## Tech Stack

Go, Fiber, gRPC, SQLC, PostgreSQL, Redis, NATS JetStream, Traefik, OpenTelemetry, Grafana/Tempo/Loki, buf, goose, mockery, testcontainers, go-playground/validator
