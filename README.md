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

| Service           | URL                             |
| ----------------- | ------------------------------- |
| Web UI            | http://booker.localhost          |
| API (Traefik)     | http://api.booker.localhost      |
| API Docs          | http://localhost:8090/docs       |
| Users             | http://localhost:8081            |
| Wallet            | http://localhost:8082            |
| Orders            | http://localhost:8083            |
| Matching          | gRPC :50054                     |
| Market            | http://localhost:8085            |
| Notification      | http://localhost:8086            |
| Grafana           | http://localhost:3000            |
| Traefik Dashboard | http://localhost:8888            |
| NATS Monitoring   | http://localhost:8222            |

## API Endpoints

### Auth (public)

```
POST /api/v1/auth/register  { email, password }
POST /api/v1/auth/login     { email, password }
POST /api/v1/auth/refresh   (uses HTTP-only cookie)
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
POST /api/v1/wallet/deposit    { assetId, amount }
POST /api/v1/wallet/withdraw   { assetId, amount }
GET  /api/v1/wallet
GET  /api/v1/wallet/:asset_id
```

### Orders (protected)

```
POST   /api/v1/orders          { pairId, side, price, quantity }
GET    /api/v1/orders          ?pair_id=&status=&limit=20&offset=0
GET    /api/v1/orders/:id
DELETE /api/v1/orders/:id
```

### Market (public)

```
GET /api/v1/market/pairs
GET /api/v1/market/ticker
GET /api/v1/market/ticker/:pair
GET /api/v1/market/trades/:pair    ?limit=50
GET /api/v1/market/orderbook/:pair ?depth=20
WS  /ws                            (real-time tickers + trades)
```

### Notifications (protected)

```
GET   /api/v1/notifications            ?cursor=&limit=20&only_unread=
GET   /api/v1/notifications/unread-count
PATCH /api/v1/notifications/:id/read
POST  /api/v1/notifications/read-all
WS    /api/v1/notifications/ws         (real-time notifications)
```

### Response format

```json
{
  "data": { ... },
  "traceId": "abc123",
  "requestId": "uuid"
}
```

All JSON fields use **camelCase**. Query parameters use **snake_case**.

## Development

```bash
make format         # Format code
make test           # Run all tests
make test-unit      # Unit tests only
make test-coverage  # Coverage report
make mock           # Regenerate mocks
make sqlc-generate  # Regenerate SQLC
make proto-generate # Regenerate protobuf

# API Client Generation (OpenAPI → TypeScript)
go run . openapi-export      # Export OpenAPI 3.0 spec to docs/openapi.yaml
cd web && pnpm generate:api  # Generate TS types + React Query hooks + Zod schemas
```

### API Client Pipeline

```
Go structs (required:"true" tags)
  → oaswrap/spec + fiberopenapi → OpenAPI 3.0.3
  → go run . openapi-export → docs/openapi.yaml
  → orval v8 → TypeScript interfaces + React Query hooks + Zod schemas
```

Each service also serves interactive API docs at `/docs` (Stoplight Elements).

## Tech Stack

**Backend:** Go 1.26, Fiber, gRPC, SQLC, PostgreSQL, Redis, NATS JetStream, Traefik, OpenTelemetry, Grafana/Tempo/Loki, buf, goose, mockery, testcontainers, oaswrap/spec

**Frontend:** React 19, Vite 8, TypeScript, TanStack React Query, Axios, Zod 4, Tailwind CSS, shadcn, orval

**Testing:** Go unit/integration tests, Cucumber + Playwright E2E
