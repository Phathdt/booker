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

| Service           | URL                   |
| ----------------- | --------------------- |
| API (Traefik)     | http://localhost      |
| Users             | http://localhost:8081 |
| Wallet            | http://localhost:8082 |
| Orders            | http://localhost:8083 |
| Matching          | gRPC :50054           |
| Market            | http://localhost:8085 |
| Notification      | http://localhost:8086 |
| Grafana           | http://localhost:3000 |
| Traefik Dashboard | http://localhost:8888 |
| NATS Monitoring   | http://localhost:8222 |

## API Endpoints

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

### Wallet (protected)

```
POST /api/v1/wallet/deposit    { asset_id, amount }
POST /api/v1/wallet/withdraw   { asset_id, amount }
GET  /api/v1/wallet/balances
GET  /api/v1/wallet/balances/:asset_id
```

### Orders (protected)

```
POST   /api/v1/orders          { pair_id, side, price, quantity }
GET    /api/v1/orders          ?pair_id=&status=&limit=20&offset=0
GET    /api/v1/orders/:id
DELETE /api/v1/orders/:id
```

### Market (public)

```
GET /api/v1/market/pairs
GET /api/v1/market/tickers
GET /api/v1/market/trades/:pair_id  ?limit=50
WS  /api/v1/market/ws              (real-time tickers + trades)
```

### Notifications (protected)

```
GET   /api/v1/notifications           ?limit=20&offset=0
GET   /api/v1/notifications/unread
PATCH /api/v1/notifications/:id/read
PATCH /api/v1/notifications/read-all
WS    /api/v1/notifications/ws        (real-time notifications)
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
