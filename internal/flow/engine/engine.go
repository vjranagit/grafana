package engine

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/vjranagit/grafana/internal/flow/component"
	"golang.org/x/sync/errgroup"
)

type Config struct {
	LogLevel   string
	Components []component.Config
}

type Engine struct {
	cfg        *Config
	components []component.Component
	graph      *Graph
}

func New(cfg *Config) (*Engine, error) {
	eng := &Engine{
		cfg:   cfg,
		graph: NewGraph(),
	}

	// Build component graph
	if err := eng.buildGraph(); err != nil {
		return nil, fmt.Errorf("failed to build component graph: %w", err)
	}

	return eng, nil
}

func (e *Engine) buildGraph() error {
	// TODO: Parse HCL config and instantiate components
	// For now, return empty graph
	return nil
}

func (e *Engine) Run(ctx context.Context) error {
	slog.Info("starting flow engine", "components", len(e.components))

	// Topological sort to determine component start order
	startOrder, err := e.graph.TopologicalSort()
	if err != nil {
		return fmt.Errorf("failed to sort components: %w", err)
	}

	// Start components in order
	g, ctx := errgroup.WithContext(ctx)
	var mu sync.Mutex
	startedComponents := make([]component.Component, 0, len(startOrder))

	for _, nodeID := range startOrder {
		comp := e.graph.GetComponent(nodeID)
		if comp == nil {
			continue
		}

		mu.Lock()
		startedComponents = append(startedComponents, comp)
		mu.Unlock()

		g.Go(func() error {
			slog.Debug("starting component", "id", comp.ID())
			if err := comp.Run(ctx); err != nil {
				return fmt.Errorf("component %s failed: %w", comp.ID(), err)
			}
			return nil
		})
	}

	// Wait for shutdown or error
	if err := g.Wait(); err != nil {
		slog.Error("engine error", "error", err)
		return err
	}

	slog.Info("flow engine stopped")
	return nil
}

// Graph represents the component dependency graph
type Graph struct {
	nodes      map[string]*Node
	components map[string]component.Component
	mu         sync.RWMutex
}

type Node struct {
	ID       string
	DependsOn []string
}

func NewGraph() *Graph {
	return &Graph{
		nodes:      make(map[string]*Node),
		components: make(map[string]component.Component),
	}
}

func (g *Graph) AddNode(id string, dependsOn []string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nodes[id] = &Node{
		ID:       id,
		DependsOn: dependsOn,
	}
}

func (g *Graph) AddComponent(id string, comp component.Component) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.components[id] = comp
}

func (g *Graph) GetComponent(id string) component.Component {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.components[id]
}

func (g *Graph) TopologicalSort() ([]string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Simple topological sort using DFS
	visited := make(map[string]bool)
	result := make([]string, 0, len(g.nodes))

	var visit func(string) error
	visit = func(id string) error {
		if visited[id] {
			return nil
		}
		visited[id] = true

		node, ok := g.nodes[id]
		if !ok {
			return nil
		}

		// Visit dependencies first
		for _, dep := range node.DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}

		result = append(result, id)
		return nil
	}

	for id := range g.nodes {
		if err := visit(id); err != nil {
			return nil, err
		}
	}

	return result, nil
}
