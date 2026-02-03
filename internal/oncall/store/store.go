package store

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func New(dsn string) (*Store, error) {
	// Parse DSN (sqlite://path/to/db.db)
	driver := "sqlite3"
	dbPath := strings.TrimPrefix(dsn, "sqlite://")

	db, err := sql.Open(driver, dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &Store{db: db}

	// Initialize schema
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return store, nil
}

func (s *Store) migrate() error {
	schema := `
		CREATE TABLE IF NOT EXISTS schedules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			timezone TEXT NOT NULL DEFAULT 'UTC',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS schedule_layers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			schedule_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			rotation_type TEXT NOT NULL, -- daily, weekly, custom
			rotation_start DATETIME NOT NULL,
			duration_hours INTEGER NOT NULL,
			users TEXT NOT NULL, -- JSON array of user IDs
			FOREIGN KEY (schedule_id) REFERENCES schedules(id)
		);

		CREATE TABLE IF NOT EXISTS escalation_chains (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS escalation_policies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			chain_id INTEGER NOT NULL,
			step_number INTEGER NOT NULL,
			policy_type TEXT NOT NULL, -- notify_user, notify_channel, wait
			target TEXT, -- user ID, channel name, or wait duration
			wait_seconds INTEGER DEFAULT 0,
			FOREIGN KEY (chain_id) REFERENCES escalation_chains(id)
		);

		CREATE TABLE IF NOT EXISTS alert_groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			fingerprint TEXT UNIQUE NOT NULL,
			status TEXT NOT NULL, -- firing, acknowledged, resolved
			severity TEXT,
			summary TEXT,
			description TEXT,
			labels TEXT, -- JSON
			annotations TEXT, -- JSON
			escalation_chain_id INTEGER,
			acknowledged_by TEXT,
			acknowledged_at DATETIME,
			resolved_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (escalation_chain_id) REFERENCES escalation_chains(id)
		);

		CREATE TABLE IF NOT EXISTS notifications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			alert_group_id INTEGER NOT NULL,
			channel TEXT NOT NULL, -- slack, email, webhook
			recipient TEXT NOT NULL,
			status TEXT NOT NULL, -- pending, sent, failed
			error TEXT,
			sent_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (alert_group_id) REFERENCES alert_groups(id)
		);

		CREATE TABLE IF NOT EXISTS integrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type TEXT NOT NULL, -- prometheus, grafana, webhook
			config TEXT NOT NULL, -- JSON
			escalation_chain_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (escalation_chain_id) REFERENCES escalation_chains(id)
		);

		CREATE INDEX IF NOT EXISTS idx_alert_groups_fingerprint ON alert_groups(fingerprint);
		CREATE INDEX IF NOT EXISTS idx_alert_groups_status ON alert_groups(status);
		CREATE INDEX IF NOT EXISTS idx_notifications_alert_group ON notifications(alert_group_id);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}
