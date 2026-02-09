package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"myapp/internal/module"
	"myapp/internal/pkg/database"
	"myapp/internal/service/auth"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
	configFile string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "api",
		Short: "MyApp API Server",
		Long:  "A RESTful API server built with Go, Echo, and FX",
	}
	
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config/config.yaml", "path to config file")
	
	// Serve command - starts the HTTP server
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP server",
		RunE:  runServe,
	}
	
	// Migrate command - runs database migrations
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		RunE:  runMigrate,
	}
	
	// Version command - displays version information
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("MyApp API Server\n")
			fmt.Printf("Version:    %s\n", version)
			fmt.Printf("Build Time: %s\n", buildTime)
		},
	}
	
	rootCmd.AddCommand(serveCmd, migrateCmd, versionCmd)
	
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runServe starts the HTTP server with all dependencies
func runServe(cmd *cobra.Command, args []string) error {
	app := fx.New(
		module.AppModule,
		fx.NopLogger, // Disable FX's default logger, we use Zap
	)
	
	startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	if err := app.Start(startCtx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}
	
	// Wait for interrupt signal
	<-app.Done()
	
	stopCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	if err := app.Stop(stopCtx); err != nil {
		return fmt.Errorf("failed to stop application: %w", err)
	}
	
	return nil
}

// runMigrate runs database migrations
func runMigrate(cmd *cobra.Command, args []string) error {
	fmt.Println("Running database migrations...")
	
	app := fx.New(
		module.AppModule,
		fx.NopLogger,
		fx.Invoke(func(dbManager *database.DatabaseManager, logger *zap.Logger) error {
			// Run auth service migrations on master database
			if err := auth.RunMigrations(dbManager.MasterDB, logger); err != nil {
				return fmt.Errorf("run auth migrations: %w", err)
			}
			
			// Run migrations on tenant database if needed
			// You can add more migrations here for other services
			
			logger.Info("All migrations completed successfully")
			return nil
		}),
	)
	
	startCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := app.Start(startCtx); err != nil {
		return fmt.Errorf("failed to start application for migration: %w", err)
	}
	
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := app.Stop(stopCtx); err != nil {
		return fmt.Errorf("failed to stop application: %w", err)
	}
	
	fmt.Println("Migrations completed successfully!")
	return nil
}
