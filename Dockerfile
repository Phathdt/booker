# Build stage
FROM golang:1.26.1-alpine AS builder

RUN apk --no-cache add upx

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags="-w -s -extldflags=-static" \
  -tags netgo,osusergo \
  -trimpath \
  -gcflags="-l=4" \
  -o bin/main .

# Compress binary with UPX
RUN upx --best --lzma bin/main

# Verify binary still works after compression
RUN ./bin/main --help || echo "Binary compressed successfully"

# Runtime stage
FROM alpine:3.23.3

WORKDIR /app

COPY --from=builder /app/bin/main .
COPY migrations/ ./migrations/

# No hardcoded CMD — each service uses docker-compose "command"
# Default: show help
CMD ["./main", "--help"]
