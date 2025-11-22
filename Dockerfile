# Build stage - Use Debian instead of Alpine (better compatibility with confluent-kafka-go)
FROM golang:1.23-bullseye AS builder

# Set working directory
WORKDIR /app

# Install build dependencies (for confluent-kafka-go C library)
RUN apt-get update && apt-get install -y \
    git \
    gcc \
    g++ \
    librdkafka-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies and ensure go.mod is up to date
RUN go mod download && go mod tidy

# Copy source code
COPY . .

# Build the application (with CGO enabled for confluent-kafka-go)
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Final stage - Use Debian slim instead of Alpine
FROM debian:bullseye-slim

# Install runtime dependencies (librdkafka for confluent-kafka-go)
RUN apt-get update && apt-get install -y \
    ca-certificates \
    librdkafka1 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]

