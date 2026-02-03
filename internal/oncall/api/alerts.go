package api

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/vjranagit/grafana/internal/oncall/models"
	"github.com/vjranagit/grafana/internal/oncall/store"
)

// PrometheusWebhook represents the AlertManager webhook format
type PrometheusWebhook struct {
	Version  string            `json:"version"`
	GroupKey string            `json:"groupKey"`
	Status   string            `json:"status"`
	Alerts   []PrometheusAlert `json:"alerts"`
}

type PrometheusAlert struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
}

// AlertProcessor handles alert ingestion and processing
type AlertProcessor struct {
	store *store.Store
}

func NewAlertProcessor(st *store.Store) *AlertProcessor {
	return &AlertProcessor{store: st}
}

// ProcessPrometheusWebhook processes Prometheus AlertManager webhook
func (p *AlertProcessor) ProcessPrometheusWebhook(webhook *PrometheusWebhook) ([]*models.AlertGroup, error) {
	var alertGroups []*models.AlertGroup

	for _, alert := range webhook.Alerts {
		fingerprint := generateFingerprint(alert.Labels)

		severity := alert.Labels["severity"]
		if severity == "" {
			severity = "info"
		}

		summary := alert.Annotations["summary"]
		if summary == "" {
			summary = alert.Labels["alertname"]
		}

		description := alert.Annotations["description"]

		labelsJSON, _ := json.Marshal(alert.Labels)
		annotationsJSON, _ := json.Marshal(alert.Annotations)

		alertGroup := &models.AlertGroup{
			Fingerprint: fingerprint,
			Status:      alert.Status,
			Severity:    severity,
			Summary:     summary,
			Description: description,
			Labels:      alert.Labels,
			Annotations: alert.Annotations,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Store or update alert in database
		if err := p.upsertAlert(alertGroup, labelsJSON, annotationsJSON); err != nil {
			return nil, fmt.Errorf("failed to store alert: %w", err)
		}

		alertGroups = append(alertGroups, alertGroup)
	}

	return alertGroups, nil
}

// generateFingerprint creates a unique fingerprint from alert labels
func generateFingerprint(labels map[string]string) string {
	// Sort labels for consistent fingerprinting
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		// Skip certain labels that don't define alert identity
		if k == "severity" || strings.HasPrefix(k, "__") {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%s", k, labels[k]))
	}

	fingerprint := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(fingerprint))
	return fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes for readability
}

func (p *AlertProcessor) upsertAlert(alert *models.AlertGroup, labelsJSON, annotationsJSON []byte) error {
	query := `
		INSERT INTO alert_groups (fingerprint, status, severity, summary, description, labels, annotations, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(fingerprint) DO UPDATE SET
			status = excluded.status,
			severity = excluded.severity,
			summary = excluded.summary,
			description = excluded.description,
			labels = excluded.labels,
			annotations = excluded.annotations,
			updated_at = excluded.updated_at
		RETURNING id
	`

	err := p.store.DB().QueryRow(query,
		alert.Fingerprint,
		alert.Status,
		alert.Severity,
		alert.Summary,
		alert.Description,
		labelsJSON,
		annotationsJSON,
		alert.CreatedAt,
		alert.UpdatedAt,
	).Scan(&alert.ID)

	return err
}
