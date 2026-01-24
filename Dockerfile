# Stage 1: Builder
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy source code first because we need of missing go.sum on host
COPY . .

# Generate go.sum and download dependencies
RUN go mod tidy
RUN go mod download

# Build
RUN go build -o server cmd/server/main.go

# Stage 2: Runner
FROM alpine:latest

WORKDIR /app

# Copy binary and config
COPY --from=builder /app/server .
COPY --from=builder /app/config ./config

# Expose port
EXPOSE 8080

CMD ["./server"]
