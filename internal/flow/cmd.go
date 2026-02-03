package flow

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/vjranagit/grafana/internal/flow/engine"
)

func NewCommand() *cobra.Command {
	var configFile string
	var debug bool

	cmd := &cobra.Command{
		Use:   "flow",
		Short: "Telemetry collection agent",
		Long: `Start the telemetry collection agent for metrics, logs, and traces.
Uses component-based pipeline architecture with HCL configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Setup logging
			logLevel := slog.LevelInfo
			if debug {
				logLevel = slog.LevelDebug
			}
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: logLevel,
			}))
			slog.SetDefault(logger)

			// Load configuration
			cfg, err := loadConfig(configFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create engine
			eng, err := engine.New(cfg)
			if err != nil {
				return fmt.Errorf("failed to create engine: %w", err)
			}

			// Setup signal handling
			ctx, cancel := signal.NotifyContext(context.Background(),
				os.Interrupt, syscall.SIGTERM)
			defer cancel()

			// Start engine
			slog.Info("starting flow engine")
			if err := eng.Run(ctx); err != nil {
				return fmt.Errorf("engine error: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "flow.hcl",
		"Configuration file path")
	cmd.Flags().BoolVar(&debug, "debug", false, "Enable debug logging")

	return cmd
}

func loadConfig(path string) (*engine.Config, error) {
	// For now, return default config
	// TODO: Implement HCL parsing
	return &engine.Config{
		LogLevel: "info",
	}, nil
}
