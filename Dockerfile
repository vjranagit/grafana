# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN make build

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs tzdata

# Create app user
RUN addgroup -g 1000 grafana && \
    adduser -D -u 1000 -G grafana grafana

# Create directories
RUN mkdir -p /etc/grafana-ops /var/lib/grafana-ops && \
    chown -R grafana:grafana /etc/grafana-ops /var/lib/grafana-ops

USER grafana
WORKDIR /home/grafana

# Copy binary from builder
COPY --from=builder /build/grafana-ops /usr/local/bin/grafana-ops

# Expose ports
# 8080 - OnCall API
# 9090 - Metrics
EXPOSE 8080 9090

ENTRYPOINT ["/usr/local/bin/grafana-ops"]
CMD ["oncall", "--config", "/etc/grafana-ops/config.hcl"]
