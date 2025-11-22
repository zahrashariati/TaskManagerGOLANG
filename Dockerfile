# Build stage
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install git (needed for some Go dependencies)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies and ensure go.mod is up to date
RUN go mod download && go mod tidy

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]

