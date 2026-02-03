package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vjranagit/grafana/internal/oncall"
	"github.com/vjranagit/grafana/internal/flow"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "grafana-ops",
		Short: "Grafana Operations Toolkit - Unified OnCall and Telemetry Agent",
		Long: `grafana-ops combines on-call management and telemetry collection
in a single Go binary. It reimplements features from Grafana OnCall
and Grafana Agent with a modern, cloud-native architecture.`,
		Version: fmt.Sprintf("%s (%s)", version, commit),
	}

	// Add subcommands
	rootCmd.AddCommand(oncall.NewCommand())
	rootCmd.AddCommand(flow.NewCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
