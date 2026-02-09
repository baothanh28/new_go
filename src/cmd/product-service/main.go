package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"myapp/internal/service/product"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "product-service",
		Short: "Product Service - Standalone",
		Long:  "Product service that can run independently",
	}

	// Serve command - starts the product service
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the product service",
		RunE:  runServe,
	}

	// Version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Product Service\n")
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

// runServe starts the product service with all its dependencies
func runServe(cmd *cobra.Command, args []string) error {
	app := fx.New(
		product.AppModule, // Uses product's own app.go
		fx.NopLogger,
	)

	startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.Start(startCtx); err != nil {
		return fmt.Errorf("failed to start product service: %w", err)
	}

	// Wait for interrupt signal
	<-app.Done()

	stopCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.Stop(stopCtx); err != nil {
		return fmt.Errorf("failed to stop product service: %w", err)
	}

	return nil
}
