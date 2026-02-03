package component

import (
	"context"
)

// Component represents a flow component (scraper, forwarder, etc.)
type Component interface {
	// ID returns the unique identifier for this component
	ID() string

	// Run starts the component. Should block until context is cancelled.
	Run(ctx context.Context) error

	// Health returns the current health status
	Health() Health
}

// Config represents component configuration
type Config struct {
	Type   string                 // e.g., "prometheus.scrape"
	Name   string                 // Instance name
	Config map[string]interface{} // Type-specific config
}

// Health represents component health status
type Health struct {
	Status  Status
	Message string
}

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// Registry holds registered component types
type Registry struct {
	factories map[string]Factory
}

// Factory creates a new component instance
type Factory func(cfg Config) (Component, error)

var DefaultRegistry = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]Factory),
	}
}

func (r *Registry) Register(componentType string, factory Factory) {
	r.factories[componentType] = factory
}

func (r *Registry) Create(cfg Config) (Component, error) {
	factory, ok := r.factories[cfg.Type]
	if !ok {
		return nil, ErrUnknownComponent{Type: cfg.Type}
	}
	return factory(cfg)
}

type ErrUnknownComponent struct {
	Type string
}

func (e ErrUnknownComponent) Error() string {
	return "unknown component type: " + e.Type
}
