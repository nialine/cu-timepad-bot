FROM golang:1.26.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the Go binary statically with optimizations
# CGO_ENABLED=0 creates a self-contained binary (no external C library dependencies)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/app ./cmd/bot

FROM alpine:3.20

RUN adduser -D appuser
USER appuser

WORKDIR /

COPY --from=builder /app/app /app

ENTRYPOINT ["/app"]
