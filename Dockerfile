# Build Stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=local

# Download dependencies
COPY nanny-backend/go.mod nanny-backend/go.sum ./
RUN go mod download

# Build application
COPY nanny-backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# Production Stage
FROM alpine:3.19

WORKDIR /app

# Copy binary and static assets
COPY --from=builder /app/server .
COPY nanny-backend/static ./static

# Default environment variables
ENV PORT=8080

EXPOSE 8080

CMD ["./server"]
