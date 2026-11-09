# Build Stage
FROM golang:1.23-bookworm AS builder

WORKDIR /app

# Copy all source files
COPY . ./

# Auto-detect subdirectory, tidy dependencies, and compile main.go
RUN if [ -f "go.mod" ]; then \
        go mod tidy && \
        CGO_ENABLED=0 GOOS=linux go build -o /app/server main.go && \
        mkdir -p /app/static && cp -r static/* /app/static/; \
    elif [ -d "nanny-backend" ]; then \
        cd nanny-backend && \
        go mod tidy && \
        CGO_ENABLED=0 GOOS=linux go build -o /app/server main.go && \
        mkdir -p /app/static && cp -r static/* /app/static/; \
    fi

# Production Stage
FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/server .
COPY --from=builder /app/static ./static

ENV PORT=8080

EXPOSE 8080

CMD ["./server"]
