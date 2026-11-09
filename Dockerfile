# Build Stage
FROM golang:1.23-bookworm AS builder

WORKDIR /app

# Copy application source
COPY nanny-backend/ ./

# Build application directly (resolves dependencies automatically)
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# Production Stage
FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

# Copy binary and static assets
COPY --from=builder /app/server .
COPY nanny-backend/static ./static

ENV PORT=8080

EXPOSE 8080

CMD ["./server"]
