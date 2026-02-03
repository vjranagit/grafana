package models

import (
	"testing"
	"time"
)

func TestLayer_GetOnCallUser(t *testing.T) {
	tests := []struct {
		name          string
		layer         Layer
		queryTime     time.Time
		expectedUser  string
		shouldError   bool
	}{
		{
			name: "daily rotation - first user",
			layer: Layer{
				RotationType:  "daily",
				RotationStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Users:         []string{"alice", "bob", "charlie"},
			},
			queryTime:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expectedUser: "alice",
		},
		{
			name: "daily rotation - second user",
			layer: Layer{
				RotationType:  "daily",
				RotationStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Users:         []string{"alice", "bob", "charlie"},
			},
			queryTime:    time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
			expectedUser: "bob",
		},
		{
			name: "daily rotation - wraps around",
			layer: Layer{
				RotationType:  "daily",
				RotationStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Users:         []string{"alice", "bob", "charlie"},
			},
			queryTime:    time.Date(2024, 1, 4, 12, 0, 0, 0, time.UTC),
			expectedUser: "alice",
		},
		{
			name: "weekly rotation - first week",
			layer: Layer{
				RotationType:  "weekly",
				RotationStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Users:         []string{"alice", "bob"},
			},
			queryTime:    time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC),
			expectedUser: "alice",
		},
		{
			name: "weekly rotation - second week",
			layer: Layer{
				RotationType:  "weekly",
				RotationStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Users:         []string{"alice", "bob"},
			},
			queryTime:    time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC),
			expectedUser: "bob",
		},
		{
			name: "custom rotation - 12 hour shifts",
			layer: Layer{
				RotationType:  "custom",
				RotationStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				DurationHours: 12,
				Users:         []string{"alice", "bob"},
			},
			queryTime:    time.Date(2024, 1, 1, 18, 0, 0, 0, time.UTC),
			expectedUser: "bob",
		},
		{
			name: "empty users list",
			layer: Layer{
				RotationType:  "daily",
				RotationStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Users:         []string{},
			},
			queryTime:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expectedUser: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := tt.layer.GetOnCallUser(tt.queryTime)
			if tt.shouldError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if user != tt.expectedUser {
				t.Errorf("expected user %q, got %q", tt.expectedUser, user)
			}
		})
	}
}

func TestSchedule_GetCurrentOnCall(t *testing.T) {
	queryTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	schedule := Schedule{
		ID:   1,
		Name: "Platform Team",
		Layers: []Layer{
			{
				ID:            1,
				RotationType:  "weekly",
				RotationStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Users:         []string{"alice", "bob", "charlie"},
			},
		},
	}

	user, err := schedule.GetCurrentOnCall(queryTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Jan 15 is 2 weeks after Jan 1
	// 2 weeks = 14 days, 14 / 7 = 2 rotations
	// Index 2 % 3 = 2, so should be "charlie"
	expectedUser := "charlie"
	if user != expectedUser {
		t.Errorf("expected user %q, got %q", expectedUser, user)
	}
}

func TestSchedule_GetCurrentOnCall_NoLayers(t *testing.T) {
	schedule := Schedule{
		ID:     1,
		Name:   "Empty Schedule",
		Layers: []Layer{},
	}

	user, err := schedule.GetCurrentOnCall(time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user != "" {
		t.Errorf("expected empty user, got %q", user)
	}
}
