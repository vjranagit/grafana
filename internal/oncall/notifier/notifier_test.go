package notifier

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vjranagit/grafana/internal/oncall/models"
)

func TestSlackNotifier_buildSlackMessage(t *testing.T) {
	notifier := NewSlackNotifier("https://hooks.slack.com/test")

	tests := []struct {
		name          string
		alert         *models.AlertGroup
		expectedColor string
		expectedIcon  string
	}{
		{
			name: "critical firing alert",
			alert: &models.AlertGroup{
				Fingerprint: "abc123",
				Status:      "firing",
				Severity:    "critical",
				Summary:     "High error rate detected",
				Description: "Error rate is above 5%",
				Labels: map[string]string{
					"alertname": "HighErrorRate",
					"instance":  "server1",
					"job":       "api",
				},
			},
			expectedColor: "#FF0000",
			expectedIcon:  "🔥",
		},
		{
			name: "warning alert",
			alert: &models.AlertGroup{
				Fingerprint: "def456",
				Status:      "firing",
				Severity:    "warning",
				Summary:     "High latency",
				Labels: map[string]string{
					"alertname": "HighLatency",
				},
			},
			expectedColor: "#FFA500",
			expectedIcon:  "🔥",
		},
		{
			name: "resolved alert",
			alert: &models.AlertGroup{
				Fingerprint: "ghi789",
				Status:      "resolved",
				Severity:    "critical",
				Summary:     "Issue resolved",
			},
			expectedColor: "#00FF00",
			expectedIcon:  "✅",
		},
		{
			name: "acknowledged alert",
			alert: &models.AlertGroup{
				Fingerprint: "jkl012",
				Status:      "acknowledged",
				Severity:    "warning",
				Summary:     "Alert acknowledged",
			},
			expectedColor: "#FFFF00",
			expectedIcon:  "👀",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := notifier.buildSlackMessage(tt.alert)

			if msg == nil {
				t.Fatal("expected message, got nil")
			}

			if msg.Text == "" {
				t.Error("expected non-empty text")
			}

			if len(msg.Attachments) == 0 {
				t.Fatal("expected at least one attachment")
			}

			attachment := msg.Attachments[0]
			if attachment.Color != tt.expectedColor {
				t.Errorf("expected color %s, got %s", tt.expectedColor, attachment.Color)
			}

			// Check that key fields are present
			foundStatus := false
			foundSeverity := false
			for _, field := range attachment.Fields {
				if field.Title == "Status" {
					foundStatus = true
					if field.Value != tt.alert.Status {
						t.Errorf("expected status %s, got %s", tt.alert.Status, field.Value)
					}
				}
				if field.Title == "Severity" {
					foundSeverity = true
					if field.Value != tt.alert.Severity {
						t.Errorf("expected severity %s, got %s", tt.alert.Severity, field.Value)
					}
				}
			}

			if !foundStatus {
				t.Error("status field not found")
			}
			if !foundSeverity {
				t.Error("severity field not found")
			}
		})
	}
}

func TestSlackNotifier_Send(t *testing.T) {
	// Create a test server to receive webhook
	receivedPayload := make(chan *SlackMessage, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected application/json, got %s", contentType)
		}

		body, _ := io.ReadAll(r.Body)
		var msg SlackMessage
		json.Unmarshal(body, &msg)
		receivedPayload <- &msg

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewSlackNotifier(server.URL)

	alert := &models.AlertGroup{
		ID:          1,
		Fingerprint: "test123",
		Status:      "firing",
		Severity:    "critical",
		Summary:     "Test alert",
		Description: "This is a test",
		Labels: map[string]string{
			"alertname": "TestAlert",
		},
		CreatedAt: time.Now(),
	}

	ctx := context.Background()
	err := notifier.Send(ctx, alert, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check received payload
	select {
	case msg := <-receivedPayload:
		if msg.Text == "" {
			t.Error("expected non-empty message text")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for webhook")
	}
}

func TestSlackNotifier_Send_Failure(t *testing.T) {
	// Create a test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	notifier := NewSlackNotifier(server.URL)

	alert := &models.AlertGroup{
		Fingerprint: "test123",
		Status:      "firing",
		Severity:    "critical",
		Summary:     "Test alert",
	}

	ctx := context.Background()
	err := notifier.Send(ctx, alert, "")
	if err == nil {
		t.Fatal("expected error for failed webhook, got nil")
	}
}

func TestWebhookNotifier_Send(t *testing.T) {
	receivedPayload := make(chan map[string]interface{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		json.Unmarshal(body, &payload)
		receivedPayload <- payload

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewWebhookNotifier("10s")

	alert := &models.AlertGroup{
		ID:          1,
		Fingerprint: "webhook123",
		Status:      "firing",
		Severity:    "warning",
		Summary:     "Test webhook alert",
		Labels: map[string]string{
			"test": "label",
		},
	}

	ctx := context.Background()
	err := notifier.Send(ctx, alert, server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check received payload
	select {
	case payload := <-receivedPayload:
		if payload["fingerprint"] != "webhook123" {
			t.Errorf("expected fingerprint webhook123, got %v", payload["fingerprint"])
		}
		if payload["status"] != "firing" {
			t.Errorf("expected status firing, got %v", payload["status"])
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for webhook")
	}
}

func TestManager_Register_and_Send(t *testing.T) {
	manager := NewManager()

	// Register a test notifier
	testNotifier := &mockNotifier{
		channel: "test",
		sendFn: func(ctx context.Context, alert *models.AlertGroup, recipient string) error {
			return nil
		},
	}

	manager.Register(testNotifier)

	alert := &models.AlertGroup{
		Fingerprint: "test",
		Severity:    "info",
		Summary:     "Test",
	}

	ctx := context.Background()
	err := manager.Send(ctx, "test", alert, "recipient")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test unknown channel
	err = manager.Send(ctx, "unknown", alert, "recipient")
	if err == nil {
		t.Fatal("expected error for unknown channel")
	}
}

// Mock notifier for testing
type mockNotifier struct {
	channel string
	sendFn  func(ctx context.Context, alert *models.AlertGroup, recipient string) error
}

func (m *mockNotifier) Channel() string {
	return m.channel
}

func (m *mockNotifier) Send(ctx context.Context, alert *models.AlertGroup, recipient string) error {
	return m.sendFn(ctx, alert, recipient)
}
