package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vjranagit/grafana/internal/oncall/api"
	"github.com/vjranagit/grafana/internal/oncall/store"
)

type Config struct {
	Listen   string
	Database string
}

type Server struct {
	cfg    *Config
	router *chi.Mux
	store  *store.Store
}

func New(cfg *Config) (*Server, error) {
	// Initialize database
	st, err := store.New(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize store: %w", err)
	}

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API routes
	r.Mount("/api/v1", api.NewRouter(st))

	return &Server{
		cfg:    cfg,
		router: r,
		store:  st,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:    s.cfg.Listen,
		Handler: s.router,
	}

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		slog.Info("server listening", "addr", s.cfg.Listen)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		slog.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
