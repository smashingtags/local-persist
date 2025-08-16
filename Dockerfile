# Build stage
FROM golang:1.18-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /src

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Dependencies should already be working

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o local-persist .

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies and create non-root user
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -g 1000 localvol \
    && adduser -D -s /bin/sh -u 1000 -G localvol localvol

# Create required directories with proper permissions
RUN mkdir -p /var/lib/docker/plugin-data/ \
    && chown -R localvol:localvol /var/lib/docker/plugin-data/

# Copy binary from builder stage
COPY --from=builder /src/local-persist /usr/local/bin/local-persist
RUN chmod +x /usr/local/bin/local-persist

# Switch to non-root user
USER localvol

# Health check to verify plugin socket is responding
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD test -S /run/docker/plugins/local-persist.sock || exit 1

# Expose the plugin socket directory
VOLUME ["/run/docker/plugins", "/var/lib/docker/plugin-data"]

CMD ["/usr/local/bin/local-persist"]
