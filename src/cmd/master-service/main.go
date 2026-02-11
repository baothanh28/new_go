package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"myapp/internal/service/master"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
	configFile string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "master-service",
		Short: "Master Service - Standalone",
		Long:  "Master service that includes auth and can run independently",
	}
	
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config/config.yaml", "path to config file")

	// Serve command - starts the master service
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the master service",
		RunE:  runServe,
	}

	// Version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Master Service\n")
			fmt.Printf("Version:    %s\n", version)
			fmt.Printf("Build Time: %s\n", buildTime)
		},
	}

	rootCmd.AddCommand(serveCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runServe starts the master service with all its dependencies
func runServe(cmd *cobra.Command, args []string) error {
	app := fx.New(
		master.AppModule, // Uses master's own app.go (includes auth module)
		fx.NopLogger,
	)

	startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.Start(startCtx); err != nil {
		return fmt.Errorf("failed to start master service: %w", err)
	}

	// Wait for interrupt signal
	<-app.Done()

	stopCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.Stop(stopCtx); err != nil {
		return fmt.Errorf("failed to stop master service: %w", err)
	}

	return nil
}
