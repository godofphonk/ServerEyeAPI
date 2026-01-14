# Build stage
FROM golang:1.25-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Install Wire and generate dependencies
RUN go install github.com/google/wire/cmd/wire@latest
RUN go generate ./internal/wire

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/servereye-api ./cmd/api

# Final stage
FROM alpine:latest

# Install ca-certificates and curl
RUN apk --no-cache add ca-certificates tzdata curl

# Create user
RUN addgroup -g 1001 -S servereye && \
    adduser -u 1001 -S servereye -G servereye

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/servereye-api .

# Copy .env.example as template
COPY .env.example .env.example

# Create logs directory
RUN mkdir -p /app/logs && chown -R servereye:servereye /app

# Switch to non-root user
USER servereye

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
CMD ["./servereye-api"]
