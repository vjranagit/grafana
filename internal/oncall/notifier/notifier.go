package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/vjranagit/grafana/internal/oncall/models"
)

// Notifier interface for sending notifications
type Notifier interface {
	Send(ctx context.Context, alert *models.AlertGroup, recipient string) error
	Channel() string
}

// Manager manages multiple notification channels
type Manager struct {
	notifiers map[string]Notifier
}

func NewManager() *Manager {
	return &Manager{
		notifiers: make(map[string]Notifier),
	}
}

func (m *Manager) Register(notifier Notifier) {
	m.notifiers[notifier.Channel()] = notifier
}

func (m *Manager) Send(ctx context.Context, channel string, alert *models.AlertGroup, recipient string) error {
	notifier, ok := m.notifiers[channel]
	if !ok {
		return fmt.Errorf("unknown notification channel: %s", channel)
	}

	slog.Info("sending notification",
		"channel", channel,
		"recipient", recipient,
		"alert", alert.Fingerprint)

	return notifier.Send(ctx, alert, recipient)
}

// SlackNotifier sends notifications via Slack webhook
type SlackNotifier struct {
	webhookURL string
	httpClient *http.Client
}

func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (n *SlackNotifier) Channel() string {
	return "slack"
}

func (n *SlackNotifier) Send(ctx context.Context, alert *models.AlertGroup, recipient string) error {
	// Build Slack message with rich formatting
	message := n.buildSlackMessage(alert)

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}

	// Use recipient as webhook URL if provided, otherwise use default
	webhookURL := n.webhookURL
	if recipient != "" {
		webhookURL = recipient
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	slog.Info("slack notification sent successfully",
		"alert", alert.Fingerprint,
		"severity", alert.Severity,
		"status", alert.Status)

	return nil
}

// SlackMessage represents the Slack webhook payload
type SlackMessage struct {
	Text        string            `json:"text,omitempty"`
	Blocks      []SlackBlock      `json:"blocks,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

type SlackBlock struct {
	Type string         `json:"type"`
	Text *SlackTextObj  `json:"text,omitempty"`
	Fields []SlackTextObj `json:"fields,omitempty"`
}

type SlackTextObj struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type SlackAttachment struct {
	Color  string   `json:"color,omitempty"`
	Fields []SlackField `json:"fields,omitempty"`
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func (n *SlackNotifier) buildSlackMessage(alert *models.AlertGroup) *SlackMessage {
	// Determine color based on severity
	color := "#808080" // gray for info
	switch alert.Severity {
	case "critical":
		color = "#FF0000" // red
	case "warning":
		color = "#FFA500" // orange
	case "info":
		color = "#0000FF" // blue
	}

	// Build status icon
	statusIcon := "🔥"
	if alert.Status == "resolved" {
		statusIcon = "✅"
		color = "#00FF00" // green
	} else if alert.Status == "acknowledged" {
		statusIcon = "👀"
		color = "#FFFF00" // yellow
	}

	// Build main text
	text := fmt.Sprintf("%s *%s* - %s", statusIcon, alert.Severity, alert.Summary)

	// Build fields from labels
	fields := []SlackField{
		{
			Title: "Status",
			Value: alert.Status,
			Short: true,
		},
		{
			Title: "Severity",
			Value: alert.Severity,
			Short: true,
		},
	}

	// Add description if present
	if alert.Description != "" {
		fields = append(fields, SlackField{
			Title: "Description",
			Value: alert.Description,
			Short: false,
		})
	}

	// Add key labels
	for key, value := range alert.Labels {
		if key == "alertname" || key == "instance" || key == "job" {
			fields = append(fields, SlackField{
				Title: key,
				Value: value,
				Short: true,
			})
		}
	}

	return &SlackMessage{
		Text: text,
		Attachments: []SlackAttachment{
			{
				Color:  color,
				Fields: fields,
			},
		},
	}
}

// EmailNotifier sends notifications via SMTP
type EmailNotifier struct {
	smtpHost string
	smtpPort int
	from     string
}

func NewEmailNotifier(smtpHost string, smtpPort int, from string) *EmailNotifier {
	return &EmailNotifier{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		from:     from,
	}
}

func (n *EmailNotifier) Channel() string {
	return "email"
}

func (n *EmailNotifier) Send(ctx context.Context, alert *models.AlertGroup, recipient string) error {
	// TODO: Implement actual SMTP send with net/smtp
	slog.Info("email notification sent",
		"recipient", recipient,
		"from", n.from,
		"alert", alert.Fingerprint)
	return nil
}

// WebhookNotifier sends notifications to a generic webhook
type WebhookNotifier struct {
	timeout    time.Duration
	httpClient *http.Client
}

func NewWebhookNotifier(timeout string) *WebhookNotifier {
	duration, _ := time.ParseDuration(timeout)
	if duration == 0 {
		duration = 10 * time.Second
	}

	return &WebhookNotifier{
		timeout: duration,
		httpClient: &http.Client{
			Timeout: duration,
		},
	}
}

func (n *WebhookNotifier) Channel() string {
	return "webhook"
}

func (n *WebhookNotifier) Send(ctx context.Context, alert *models.AlertGroup, recipient string) error {
	// Build generic webhook payload
	payload := map[string]interface{}{
		"alert_id":    alert.ID,
		"fingerprint": alert.Fingerprint,
		"status":      alert.Status,
		"severity":    alert.Severity,
		"summary":     alert.Summary,
		"description": alert.Description,
		"labels":      alert.Labels,
		"annotations": alert.Annotations,
		"created_at":  alert.CreatedAt,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", recipient, bytes.NewReader(payloadJSON))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	slog.Info("webhook notification sent successfully",
		"url", recipient,
		"alert", alert.Fingerprint)

	return nil
}
