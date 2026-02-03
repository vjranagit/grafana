package models

import (
	"time"
)

// Schedule represents an on-call schedule
type Schedule struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Timezone    string    `json:"timezone"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Layers      []Layer   `json:"layers,omitempty"`
}

// Layer represents a schedule layer (rotation)
type Layer struct {
	ID             int64     `json:"id"`
	ScheduleID     int64     `json:"schedule_id"`
	Name           string    `json:"name"`
	RotationType   string    `json:"rotation_type"` // daily, weekly, custom
	RotationStart  time.Time `json:"rotation_start"`
	DurationHours  int       `json:"duration_hours"`
	Users          []string  `json:"users"` // User IDs in rotation
}

// GetCurrentOnCall returns the user currently on-call for this schedule
func (s *Schedule) GetCurrentOnCall(t time.Time) (string, error) {
	// Simple rotation logic
	for _, layer := range s.Layers {
		user, err := layer.GetOnCallUser(t)
		if err == nil && user != "" {
			return user, nil
		}
	}
	return "", nil
}

// GetOnCallUser returns the on-call user for this layer at time t
func (l *Layer) GetOnCallUser(t time.Time) (string, error) {
	if len(l.Users) == 0 {
		return "", nil
	}

	// Calculate duration since rotation start
	duration := t.Sub(l.RotationStart)

	var rotationInterval time.Duration
	switch l.RotationType {
	case "daily":
		rotationInterval = 24 * time.Hour
	case "weekly":
		rotationInterval = 7 * 24 * time.Hour
	default:
		rotationInterval = time.Duration(l.DurationHours) * time.Hour
	}

	// Find current position in rotation
	rotations := int(duration / rotationInterval)
	userIndex := rotations % len(l.Users)

	return l.Users[userIndex], nil
}

// EscalationChain represents an escalation policy
type EscalationChain struct {
	ID          int64              `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	CreatedAt   time.Time          `json:"created_at"`
	Policies    []EscalationPolicy `json:"policies,omitempty"`
}

// EscalationPolicy represents a step in an escalation chain
type EscalationPolicy struct {
	ID          int64  `json:"id"`
	ChainID     int64  `json:"chain_id"`
	StepNumber  int    `json:"step_number"`
	PolicyType  string `json:"policy_type"` // notify_user, notify_channel, wait
	Target      string `json:"target"`      // user ID, channel name, or wait duration
	WaitSeconds int    `json:"wait_seconds"`
}

// AlertGroup represents a group of related alerts
type AlertGroup struct {
	ID                 int64             `json:"id"`
	Fingerprint        string            `json:"fingerprint"`
	Status             string            `json:"status"` // firing, acknowledged, resolved
	Severity           string            `json:"severity"`
	Summary            string            `json:"summary"`
	Description        string            `json:"description"`
	Labels             map[string]string `json:"labels"`
	Annotations        map[string]string `json:"annotations"`
	EscalationChainID  *int64            `json:"escalation_chain_id,omitempty"`
	AcknowledgedBy     *string           `json:"acknowledged_by,omitempty"`
	AcknowledgedAt     *time.Time        `json:"acknowledged_at,omitempty"`
	ResolvedAt         *time.Time        `json:"resolved_at,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

// Notification represents a notification sent for an alert
type Notification struct {
	ID           int64      `json:"id"`
	AlertGroupID int64      `json:"alert_group_id"`
	Channel      string     `json:"channel"` // slack, email, webhook
	Recipient    string     `json:"recipient"`
	Status       string     `json:"status"` // pending, sent, failed
	Error        *string    `json:"error,omitempty"`
	SentAt       *time.Time `json:"sent_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Integration represents an alert source integration
type Integration struct {
	ID                 int64             `json:"id"`
	Name               string            `json:"name"`
	Type               string            `json:"type"` // prometheus, grafana, webhook
	Config             map[string]string `json:"config"`
	EscalationChainID  *int64            `json:"escalation_chain_id,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
}
