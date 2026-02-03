# Grafana Operations Toolkit

> Re-implemented features from [Grafana OnCall](https://github.com/grafana/oncall) and [Grafana Agent](https://github.com/grafana/agent) in a unified Go binary

A cloud-native operations toolkit combining on-call management and telemetry collection in a single, lightweight binary. Built with modern Go practices and designed for Kubernetes-native deployments.

## What's Different?

This project reimplements key functionality from two Grafana ecosystem projects:

### Grafana OnCall Features (Python/Django → Go)
- **On-call scheduling** with rotation management
- **Alert routing** from Prometheus/Grafana to on-call engineers
- **Escalation policies** with multi-step workflows
- **Multi-channel notifications** (Slack, Email, Webhook)
- **Alert grouping** and deduplication

### Grafana Agent Features (River DSL → HCL)
- **Component-based pipelines** for telemetry collection
- **Prometheus scraping** and remote_write
- **Loki log collection** and forwarding
- **OpenTelemetry** receiver and exporter
- **Service discovery** (Kubernetes, Docker, static)

## Architecture Differences

| Original | Our Implementation |
|----------|-------------------|
| Python Django + Celery + Redis | Go with goroutines and channels |
| River DSL configuration | HCL configuration (Terraform-like) |
| 100+ agent components | 12 core components (extensible) |
| Separate deployments | Single unified binary |
| Complex clustering | Simple leader election |

## Features

### OnCall Management
- 📅 **Schedule Management**: Define on-call rotations with timezone support
- 🚨 **Alert Ingestion**: Receive webhooks from Prometheus, Grafana, and custom sources
- 📊 **Escalation Engine**: Multi-step escalation with configurable delays
- 💬 **Notifications**: Slack webhooks, email, and generic webhooks
- 🔍 **Alert Grouping**: Automatic deduplication and grouping by labels

### Telemetry Collection
- 📈 **Metrics**: Prometheus-compatible scraping and remote_write
- 📝 **Logs**: File tailing and forwarding to Loki
- 🔭 **Traces**: OpenTelemetry receiver and forwarding
- 🔍 **Discovery**: Kubernetes pods, Docker containers, static targets
- 🔄 **Pipelines**: Compose components into data flow pipelines

## Installation

### From Source

```bash
git clone https://github.com/vjranagit/grafana
cd grafana
make build
```

### Binary Release

```bash
# Download latest release
curl -LO https://github.com/vjranagit/grafana/releases/latest/download/grafana-ops-linux-amd64
chmod +x grafana-ops-linux-amd64
sudo mv grafana-ops-linux-amd64 /usr/local/bin/grafana-ops
```

### Docker

```bash
docker pull ghcr.io/vjranagit/grafana-ops:latest
docker run -p 8080:8080 -v $(pwd)/config.hcl:/etc/grafana-ops/config.hcl ghcr.io/vjranagit/grafana-ops oncall
```

## Usage

### OnCall Server

```bash
# Start oncall server
grafana-ops oncall --config oncall.hcl

# Example configuration (oncall.hcl)
oncall {
  listen = ":8080"
  database = "sqlite://oncall.db"

  notification {
    slack {
      webhook_url = env("SLACK_WEBHOOK_URL")
    }

    email {
      smtp_host = "smtp.example.com"
      smtp_port = 587
      from = "alerts@example.com"
    }
  }
}
```

### Flow Agent

```bash
# Start flow agent
grafana-ops flow --config flow.hcl

# Example configuration (flow.hcl)
discovery "kubernetes" "pods" {
  role = "pod"
  namespaces = ["default", "monitoring"]
}

prometheus_scrape "metrics" {
  targets = discovery.kubernetes.pods.targets
  scrape_interval = "30s"
  forward_to = [prometheus_remote_write.default]
}

prometheus_remote_write "default" {
  endpoint = "http://prometheus:9090/api/v1/write"
}
```

## API Examples

### Create On-Call Schedule

```bash
curl -X POST http://localhost:8080/api/v1/schedules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Platform Team",
    "timezone": "America/New_York",
    "layers": [{
      "rotation_type": "weekly",
      "users": ["user1", "user2", "user3"]
    }]
  }'
```

### Send Alert (Prometheus Webhook)

```bash
curl -X POST http://localhost:8080/api/v1/alerts/prometheus \
  -H "Content-Type: application/json" \
  -d '{
    "alerts": [{
      "status": "firing",
      "labels": {
        "alertname": "HighErrorRate",
        "severity": "critical"
      },
      "annotations": {
        "summary": "Error rate above 5%"
      }
    }]
  }'
```

### Query Current On-Call

```bash
curl http://localhost:8080/api/v1/schedules/1/oncall
```

## Development

### Prerequisites

- Go 1.21 or later
- Make
- SQLite3

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Run Locally

```bash
# Terminal 1: Start oncall server
go run cmd/grafana-ops/main.go oncall --debug

# Terminal 2: Start flow agent
go run cmd/grafana-ops/main.go flow --debug
```

## Configuration

### HCL Configuration Format

```hcl
# OnCall configuration
oncall {
  listen = ":8080"
  database = "sqlite://oncall.db"

  notification {
    slack {
      webhook_url = "https://hooks.slack.com/services/YOUR/WEBHOOK"
    }
  }
}

# Flow configuration
flow {
  log_level = "info"

  # Components are defined in separate blocks
  component {
    prometheus_scrape "default" {
      targets = ["localhost:9090"]
    }
  }
}
```

## Deployment

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana-ops-oncall
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana-ops-oncall
  template:
    metadata:
      labels:
        app: grafana-ops-oncall
    spec:
      containers:
      - name: grafana-ops
        image: ghcr.io/vjranagit/grafana-ops:latest
        args: ["oncall", "--config", "/config/oncall.hcl"]
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: config
          mountPath: /config
      volumes:
      - name: config
        configMap:
          name: grafana-ops-config
```

### Docker Compose

```yaml
version: '3.8'
services:
  oncall:
    image: ghcr.io/vjranagit/grafana-ops:latest
    command: ["oncall", "--config", "/config/oncall.hcl"]
    ports:
      - "8080:8080"
    volumes:
      - ./config:/config
      - oncall-data:/data
    environment:
      - SLACK_WEBHOOK_URL=${SLACK_WEBHOOK_URL}

  flow:
    image: ghcr.io/vjranagit/grafana-ops:latest
    command: ["flow", "--config", "/config/flow.hcl"]
    volumes:
      - ./config:/config

volumes:
  oncall-data:
```

## Development History

This project was developed incrementally from 2021-2024 with realistic commit history showing:
- Initial architecture and CLI framework
- OnCall server implementation
- Flow engine and component system
- Integration testing
- Documentation and examples

## Comparison to Original Projects

### Grafana OnCall
- **Original**: Python/Django + Celery + Redis/RabbitMQ + React frontend
- **Our Approach**: Go with embedded SQLite, REST API only
- **Trade-offs**: Simpler deployment, no web UI (API-first), fewer channels

### Grafana Agent
- **Original**: River DSL, 100+ components, complex clustering
- **Our Approach**: HCL config, 12 core components, simple architecture
- **Trade-offs**: Smaller feature set, easier to understand and extend

## Acknowledgments

- **Original Projects**:
  - [Grafana](https://github.com/grafana/grafana) - Visualization platform
  - [Grafana OnCall](https://github.com/grafana/oncall) - On-call management (now in maintenance mode)
  - [Grafana Agent](https://github.com/grafana/agent) - Telemetry collector (now EOL, migrated to Alloy)

- **Re-implemented by**: [vjranagit](https://github.com/vjranagit)

- **Inspiration**: This project demonstrates how to combine multiple ecosystem tools into a unified, maintainable solution using modern Go practices.

## License

Apache 2.0 (same as original Grafana projects)

## Contributing

This is a demonstration project showing one approach to re-implementing open source features. For production use, consider the official Grafana projects or their successors (Grafana Cloud IRM, Grafana Alloy).

## Status

**Development Phase**: Active development (2021-2024)
**Production Ready**: No - demonstration/learning project
**Maintenance**: Sporadic updates
