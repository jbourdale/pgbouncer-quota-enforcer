package interfaces

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"pgbouncer-quota-enforcer/internal/app"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// NewServerCommand creates the server command
func NewServerCommand() *cobra.Command {
	var address string

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the TCP server that logs received bytes",
		Long: `Start a TCP server that accepts connections and logs all received bytes.
This server is designed to be the first step in building a PostgreSQL
protocol-aware quota enforcement service.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(address)
		},
	}

	cmd.Flags().StringVarP(&address, "address", "a", ":5432", "Address to listen on (default: :5432)")

	return cmd
}

// runServer starts the TCP server and handles graceful shutdown
func runServer(address string) error {
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create server service
	serverService := app.NewServerService(app.ServerConfig{
		Address: address,
	})

	// Start server
	if err := serverService.Start(ctx, address); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	fmt.Printf("TCP server started on %s\n", serverService.Address())
	fmt.Println("Press Ctrl+C to stop the server")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal
	<-sigChan
	fmt.Println("\nShutting down server...")

	// Create context with timeout for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Stop server
	if err := serverService.Stop(shutdownCtx); err != nil {
		return fmt.Errorf("error during server shutdown: %w", err)
	}

	fmt.Println("Server stopped successfully")
	return nil
}

// NewRootCommand creates the root command
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pgbouncer-quota-enforcer",
		Short: "PgBouncer Quota Enforcement Service",
		Long: `A high-performance quota enforcement service that processes PostgreSQL 
query events to track and enforce database usage quotas.`,
	}

	// Add subcommands
	cmd.AddCommand(NewServerCommand())

	return cmd
}
