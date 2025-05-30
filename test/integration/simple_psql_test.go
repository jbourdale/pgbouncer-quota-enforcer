//go:build integration
// +build integration

package integration

import (
	"context"
	"os/exec"
	"pgbouncer-quota-enforcer/internal/app"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// isPsqlAvailable checks if psql command is available
func isPsqlAvailable() bool {
	_, err := exec.LookPath("psql")
	return err == nil
}

func TestSimplePsqlConnection(t *testing.T) {
	// Check if psql is available
	if !isPsqlAvailable() {
		t.Skip("psql not available, skipping integration test")
	}

	t.Log("=== Starting PostgreSQL Integration Test ===")

	// Start the server
	serverService := app.NewServerService(app.ServerConfig{})

	// Start server in background
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	err := serverService.Start(serverCtx, ":15432")
	require.NoError(t, err, "Failed to start test server")

	// Give server time to start
	time.Sleep(200 * time.Millisecond)

	// Ensure server is stopped when test completes
	defer func() {
		t.Log("=== Stopping Server ===")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		serverService.Stop(shutdownCtx)
	}()

	t.Log("Server started on port 15432")
	t.Log("Sending PostgreSQL connection attempts...")

	// Define a set of queries to test different PostgreSQL message types
	queries := []string{
		"SELECT 1;",
		"SELECT 'Hello PostgreSQL';",
		"SELECT version();",
		"SELECT NOW();",
		"SELECT * FROM pg_catalog.pg_tables LIMIT 1;",
		"CREATE TABLE test_quota (id INTEGER, usage BIGINT);",
		"INSERT INTO test_quota VALUES (1, 1000);",
		"SELECT * FROM test_quota;",
		"UPDATE test_quota SET usage = 2000 WHERE id = 1;",
		"DELETE FROM test_quota WHERE id = 1;",
		"DROP TABLE test_quota;",
	}

	t.Logf("Will execute %d PostgreSQL queries/commands", len(queries))

	// Execute each query - psql will send PostgreSQL wire protocol data
	for i, query := range queries {
		t.Logf("--- Query %d: %s ---", i+1, query)

		// Execute psql command with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

		cmd := exec.CommandContext(ctx, "psql",
			"-h", "localhost",
			"-p", "15432",
			"-U", "testuser",
			"-d", "testdb",
			"-c", query,
		)

		// Set password to avoid prompts
		cmd.Env = append(cmd.Env, "PGPASSWORD=testpass")

		// Execute - we expect this to fail since we don't implement PG protocol
		// But it will send connection and query data to our server
		output, err := cmd.CombinedOutput()
		cancel()

		if err != nil {
			t.Logf("psql command failed (expected): %v", err)
		}

		// Log first line of output for context
		if len(output) > 0 {
			lines := strings.Split(string(output), "\n")
			if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
				t.Logf("psql output: %s", strings.TrimSpace(lines[0]))
			}
		}

		// Small delay to see individual connections in logs
		time.Sleep(100 * time.Millisecond)
	}

	t.Log("=== Integration Test Complete ===")
	t.Log("Check the server logs above to see the PostgreSQL wire protocol data received")
}
