FROM golang:1.26-alpine AS builder
RUN apk add --no-cache ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags='-w -s' -o /local-persist .

FROM alpine:3.22
RUN apk add --no-cache ca-certificates \
    && mkdir -p /var/lib/docker/plugin-data /run/docker/plugins
COPY --from=builder /local-persist /usr/local/bin/local-persist
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD test -S /run/docker/plugins/local-persist.sock || exit 1
VOLUME ["/run/docker/plugins", "/var/lib/docker/plugin-data"]
# Root required: plugin creates Unix socket at /run/docker/plugins/ and manages host directories
USER root
CMD ["/usr/local/bin/local-persist"]
