package oncall

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/vjranagit/grafana/internal/oncall/server"
)

func NewCommand() *cobra.Command {
	var configFile string
	var debug bool

	cmd := &cobra.Command{
		Use:   "oncall",
		Short: "On-call management server",
		Long: `Start the on-call management server for schedule management,
alert routing, and escalation policies.`,
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

			// Create server
			srv, err := server.New(cfg)
			if err != nil {
				return fmt.Errorf("failed to create server: %w", err)
			}

			// Setup signal handling
			ctx, cancel := signal.NotifyContext(context.Background(),
				os.Interrupt, syscall.SIGTERM)
			defer cancel()

			// Start server
			slog.Info("starting oncall server", "addr", cfg.Listen)
			if err := srv.Run(ctx); err != nil {
				return fmt.Errorf("server error: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "oncall.hcl",
		"Configuration file path")
	cmd.Flags().BoolVar(&debug, "debug", false, "Enable debug logging")

	return cmd
}

func loadConfig(path string) (*server.Config, error) {
	// For now, return default config
	// TODO: Implement HCL parsing
	return &server.Config{
		Listen:   ":8080",
		Database: "sqlite://oncall.db",
	}, nil
}
