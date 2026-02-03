# Grafana Flow Configuration Example

flow {
  # Logging configuration
  log_level = "info"

  # Component health check interval
  health_check_interval = "30s"
}

# Kubernetes pod discovery
discovery "kubernetes" "pods" {
  # Discover pods in specific namespaces
  role       = "pod"
  namespaces = ["default", "monitoring", "production"]

  # Filter by labels
  selector {
    label {
      key   = "app"
      value = ".*"
      regex = true
    }
  }
}

# Docker container discovery
discovery "docker" "containers" {
  # Docker daemon socket
  host = "unix:///var/run/docker.sock"

  # Filter by labels
  filter {
    label {
      key   = "prometheus.scrape"
      value = "true"
    }
  }
}

# Static targets
discovery "static" "manual" {
  targets = [
    {
      address = "localhost:9090"
      labels = {
        job     = "prometheus"
        env     = "dev"
      }
    },
    {
      address = "localhost:3000"
      labels = {
        job     = "grafana"
        env     = "dev"
      }
    },
  ]
}

# Prometheus scraping from Kubernetes pods
prometheus_scrape "kubernetes_pods" {
  # Targets from discovery
  targets = discovery.kubernetes.pods.targets

  # Scrape configuration
  scrape_interval = "30s"
  scrape_timeout  = "10s"
  metrics_path    = "/metrics"

  # Relabeling
  relabel_config {
    source_labels = ["__meta_kubernetes_pod_name"]
    target_label  = "pod"
  }

  relabel_config {
    source_labels = ["__meta_kubernetes_namespace"]
    target_label  = "namespace"
  }

  # Forward to remote write
  forward_to = [prometheus_remote_write.default]
}

# Prometheus remote write
prometheus_remote_write "default" {
  # Endpoint URL
  endpoint = "http://prometheus:9090/api/v1/write"

  # Queue configuration
  queue_config {
    capacity            = 10000
    max_shards          = 10
    max_samples_per_send = 1000
    batch_send_deadline = "5s"
  }

  # Authentication
  basic_auth {
    username = env("PROM_USER")
    password = env("PROM_PASS")
  }

  # Retry configuration
  retry_config {
    max_retries = 3
    min_backoff = "1s"
    max_backoff = "30s"
  }
}

# Loki log collection from files
loki_source_file "logs" {
  # File paths to tail
  paths = [
    "/var/log/app/*.log",
    "/var/log/nginx/access.log",
  ]

  # Labels to attach
  labels = {
    job  = "logs"
    host = env("HOSTNAME")
  }

  # Forward to Loki
  forward_to = [loki_write.default]
}

# Loki write endpoint
loki_write "default" {
  # Loki endpoint
  endpoint = "http://loki:3100/loki/api/v1/push"

  # Batch configuration
  batch_config {
    max_size  = 1048576 # 1MB
    max_wait  = "1s"
  }

  # Authentication
  tenant_id = "default"

  basic_auth {
    username = env("LOKI_USER")
    password = env("LOKI_PASS")
  }
}

# OpenTelemetry OTLP receiver
otelcol_receiver_otlp "default" {
  # gRPC endpoint
  grpc {
    endpoint = "0.0.0.0:4317"
  }

  # HTTP endpoint
  http {
    endpoint = "0.0.0.0:4318"
  }

  # Forward to batch processor
  forward_to = [otelcol_processor_batch.default]
}

# OpenTelemetry batch processor
otelcol_processor_batch "default" {
  # Batch size
  send_batch_size = 1000
  timeout         = "10s"

  # Forward to exporter
  forward_to = [otelcol_exporter_otlp.default]
}

# OpenTelemetry OTLP exporter
otelcol_exporter_otlp "default" {
  # Endpoint
  endpoint = "tempo:4317"

  # TLS configuration
  tls {
    insecure = true
  }

  # Retry configuration
  retry_on_failure {
    enabled         = true
    initial_interval = "5s"
    max_interval    = "30s"
    max_elapsed_time = "5m"
  }
}
