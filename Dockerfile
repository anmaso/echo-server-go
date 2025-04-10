# Build stage
FROM golang:1.24 AS builder

# Install git and SSL certificates
#RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /echo-server ./cmd/server

# Final stage
FROM alpine:3.18

# Install SSL certificates
#RUN apk add --no-cache ca-certificates

# Create non-root user
RUN adduser -D -g '' appuser

# Create configuration directories
RUN mkdir -p /app/config/paths && \
    chown -R appuser:appuser /app

# Copy binary from builder
COPY --from=builder /echo-server /app/

# Copy default configurations
COPY config/server.json /app/config/
COPY config/paths/* /app/config/paths/

# Set working directory
WORKDIR /app

# Switch to non-root user
USER appuser

# Expose default port
EXPOSE 8080

# Set environment variables
ENV CONFIG_DIR=/app/config

# Run the application
CMD ["/app/echo-server"]