# Copyright (c) 2026 godofphonk
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

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

# Force clean Go cache and rebuild with fresh modules
RUN go clean -cache -modcache && \
    rm -rf /root/.cache/go-build && \
    rm -rf /root/go/pkg/mod && \
    go mod download -x && \
    go mod verify

# Build the application with timestamp to force rebuild - use complete rebuild
ARG BUILD_DATE
ARG VERSION
ARG COMMIT_SHA
RUN rm -rf /app/servereye-api && \
    rm -rf /app/internal/websocket/*.o && \
    rm -rf /app/internal/websocket/*.a && \
    go clean -cache && \
    CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s -X main.BuildDate=${BUILD_DATE} -X main.Version=${VERSION} -X main.CommitSHA=${COMMIT_SHA}" -o /app/servereye-api ./cmd/api

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

# Copy source code for verification
COPY --from=builder /app/internal ./internal

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
