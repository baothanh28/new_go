package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"myapp/internal/service/auth"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "auth-service",
		Short: "Auth Service - Standalone",
		Long:  "Authentication service that can run independently",
	}

	// Serve command - starts the auth service
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the auth service",
		RunE:  runServe,
	}

	// Version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Auth Service\n")
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

// runServe starts the auth service with all its dependencies
func runServe(cmd *cobra.Command, args []string) error {
	app := fx.New(
		auth.AppModule, // Uses auth's own app.go
		fx.NopLogger,
	)

	startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.Start(startCtx); err != nil {
		return fmt.Errorf("failed to start auth service: %w", err)
	}

	// Wait for interrupt signal
	<-app.Done()

	stopCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.Stop(stopCtx); err != nil {
		return fmt.Errorf("failed to stop auth service: %w", err)
	}

	return nil
}
