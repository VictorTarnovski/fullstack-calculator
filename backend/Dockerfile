# Build stage
FROM golang:1.26-bookworm AS builder

WORKDIR /app

# Copy go.mod (and go.sum if it exists)
COPY go.mod .

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

# Runtime stage
FROM debian:bookworm-slim

# Install ca-certificates for HTTPS (if needed)
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/api .

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["./api"]
