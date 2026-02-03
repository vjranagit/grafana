# Grafana OnCall Configuration Example

oncall {
  # Server listen address
  listen = ":8080"

  # Database connection string
  # Supported: sqlite://, postgresql://
  database = "sqlite://oncall.db"

  # Notification channels
  notification {
    # Slack notifications via webhook
    slack {
      enabled     = true
      webhook_url = env("SLACK_WEBHOOK_URL")
      channel     = "#alerts"
      username    = "Grafana OnCall"
    }

    # Email notifications
    email {
      enabled   = true
      smtp_host = "smtp.gmail.com"
      smtp_port = 587
      smtp_user = env("SMTP_USER")
      smtp_pass = env("SMTP_PASS")
      from      = "alerts@example.com"
    }

    # Generic webhook
    webhook {
      enabled = true
      timeout = "10s"
    }
  }

  # Alert grouping settings
  grouping {
    # Group alerts with same fingerprint within window
    window        = "5m"
    # Maximum alerts per group
    max_group_size = 100
  }

  # Escalation defaults
  escalation {
    # Default wait time between escalation steps
    default_wait = "5m"
    # Maximum escalation steps
    max_steps = 10
  }
}

# Example schedule definition
schedule "platform-team" {
  name        = "Platform Team On-Call"
  description = "Primary on-call rotation for platform team"
  timezone    = "America/New_York"

  layer "primary" {
    rotation_type  = "weekly"
    rotation_start = "2024-01-01T00:00:00Z"
    users          = ["alice", "bob", "charlie"]
  }

  layer "backup" {
    rotation_type  = "weekly"
    rotation_start = "2024-01-01T00:00:00Z"
    users          = ["david", "eve"]
  }
}

# Example escalation chain
escalation_chain "critical" {
  name        = "Critical Alert Escalation"
  description = "Escalation for critical production alerts"

  step {
    type   = "notify_oncall"
    target = "platform-team"
  }

  step {
    type         = "wait"
    wait_seconds = 300 # 5 minutes
  }

  step {
    type   = "notify_channel"
    target = "slack:#incidents"
  }

  step {
    type         = "wait"
    wait_seconds = 600 # 10 minutes
  }

  step {
    type   = "notify_channel"
    target = "email:oncall-escalation@example.com"
  }
}

# Example integration
integration "prometheus" {
  name = "Prometheus AlertManager"
  type = "prometheus"

  # Link to escalation chain
  escalation_chain = "critical"

  # Filter alerts by labels
  filter {
    label_match {
      severity = "critical"
    }
  }
}
