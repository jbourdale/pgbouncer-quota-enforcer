//go:build integration
// +build integration

package integration

import (
	"context"
	"net"
	"pgbouncer-quota-enforcer/internal/app/domain"
	"pgbouncer-quota-enforcer/internal/infra/adapters"
	"pgbouncer-quota-enforcer/pkg/logger"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NormalizationTestLogger captures both original and normalized queries
type NormalizationTestLogger struct {
	queries         []string
	normalizedData  []domain.NormalizedQuery
	mu              sync.Mutex
	expectedQueries int
	queryChannel    chan string
}

func NewNormalizationTestLogger(expectedQueries int) *NormalizationTestLogger {
	return &NormalizationTestLogger{
		queries:         make([]string, 0),
		normalizedData:  make([]domain.NormalizedQuery, 0),
		expectedQueries: expectedQueries,
		queryChannel:    make(chan string, expectedQueries),
	}
}

func (t *NormalizationTestLogger) LogQuery(connectionID string, query string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.queries = append(t.queries, query)

	// Send to channel for synchronization
	select {
	case t.queryChannel <- query:
	default:
	}

	return nil
}

func (t *NormalizationTestLogger) LogNormalizedQuery(connectionID string, normalizedQuery domain.NormalizedQuery) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.normalizedData = append(t.normalizedData, normalizedQuery)
	return nil
}

func (t *NormalizationTestLogger) LogProtocolMessage(connectionID string, messageType string, details map[string]interface{}) error {
	// No-op for this test
	return nil
}

func (t *NormalizationTestLogger) GetQueries() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]string(nil), t.queries...)
}

func (t *NormalizationTestLogger) GetNormalizedData() []domain.NormalizedQuery {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]domain.NormalizedQuery(nil), t.normalizedData...)
}

func (t *NormalizationTestLogger) WaitForQueries(timeout time.Duration) []string {
	timeoutCh := time.After(timeout)
	receivedQueries := make([]string, 0, t.expectedQueries)

	for len(receivedQueries) < t.expectedQueries {
		select {
		case query := <-t.queryChannel:
			receivedQueries = append(receivedQueries, query)
		case <-timeoutCh:
			return receivedQueries
		}
	}

	return receivedQueries
}

func TestQueryNormalizationIntegration(t *testing.T) {
	t.Log("=== Starting Query Normalization Integration Test ===")

	// Create test logger to capture normalization data
	testLogger := NewNormalizationTestLogger(4)

	// Create service with our test logger
	log := logger.NewSimpleLogger()
	queryNormalizer := adapters.NewPgQueryNormalizer()
	connHandler := adapters.NewPostgreSQLConnectionHandler(testLogger, queryNormalizer, log)
	tcpServer := adapters.NewStandardTCPServer(connHandler, log)

	// Start server
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	err := tcpServer.Start(serverCtx, ":15435")
	require.NoError(t, err, "Failed to start test server")

	// Give server time to start
	time.Sleep(200 * time.Millisecond)

	// Ensure server is stopped when test completes
	defer func() {
		t.Log("=== Stopping Server ===")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		tcpServer.Stop(shutdownCtx)
	}()

	t.Log("Server started on port 15435")

	// Connect to server
	conn, err := net.Dial("tcp", "localhost:15435")
	require.NoError(t, err, "Failed to connect to test server")
	defer conn.Close()

	// Test queries to send - these should demonstrate normalization
	testQueries := []struct {
		query              string
		expectedNormalized string
		description        string
	}{
		{
			query:              "SELECT * FROM users WHERE id = 123",
			expectedNormalized: "SELECT * FROM users WHERE id = $1",
			description:        "Numeric parameter normalization",
		},
		{
			query:              "SELECT * FROM products WHERE name = 'Laptop'",
			expectedNormalized: "SELECT * FROM products WHERE name = $1",
			description:        "String parameter normalization",
		},
		{
			query:              "SELECT * FROM orders WHERE user_id = 456 AND status = 'pending'",
			expectedNormalized: "SELECT * FROM orders WHERE user_id = $1 AND status = $2",
			description:        "Multiple parameters normalization",
		},
		{
			query:              "SELECT * FROM items WHERE category IN ('electronics', 'books', 'clothing')",
			expectedNormalized: "SELECT * FROM items WHERE category IN ($1, $2, $3)",
			description:        "IN clause normalization",
		},
	}

	t.Log("Sending test queries for normalization...")

	// Send each test query
	for i, testCase := range testQueries {
		t.Logf("Sending query %d (%s): %s", i+1, testCase.description, testCase.query)

		queryMsg := &pgproto3.Query{String: testCase.query}
		queryBuf := make([]byte, 0, 1024)
		queryBuf, err = queryMsg.Encode(queryBuf)
		require.NoError(t, err, "Failed to encode query message")

		_, err = conn.Write(queryBuf)
		require.NoError(t, err, "Failed to send query message")

		// Small delay between queries
		time.Sleep(200 * time.Millisecond)
	}

	// Wait for queries to be processed
	t.Log("Waiting for queries to be processed...")
	receivedQueries := testLogger.WaitForQueries(5 * time.Second)

	// Verify all queries were received
	require.Equal(t, len(testQueries), len(receivedQueries), "All queries should be received")

	// Get normalization data
	normalizedData := testLogger.GetNormalizedData()
	require.Equal(t, len(testQueries), len(normalizedData), "All queries should have normalization data")

	// Verify each query was normalized correctly
	for i, testCase := range testQueries {
		t.Logf("Verifying normalization %d: %s", i+1, testCase.description)

		// Check original query
		assert.Equal(t, testCase.query, normalizedData[i].Original, "Original query should match")

		// Check normalized query
		assert.Equal(t, testCase.expectedNormalized, normalizedData[i].Normalized,
			"Normalized query mismatch for case: %s", testCase.description)

		// Check hash is generated
		assert.NotEmpty(t, normalizedData[i].Hash.Value(), "Query hash should be generated")

		t.Logf("✓ Query %d normalized correctly:", i+1)
		t.Logf("    Original: %s", normalizedData[i].Original)
		t.Logf("    Normalized: %s", normalizedData[i].Normalized)
		t.Logf("    Hash: %s", normalizedData[i].Hash.Value())
	}

	// Test hash consistency - queries with same structure should have same hash
	t.Log("Testing hash consistency...")

	// Send equivalent queries that should have the same hash
	equivalentQueries := []string{
		"SELECT * FROM users WHERE id = 999", // Same structure as first query
		"SELECT * FROM users WHERE id = 111", // Same structure as first query
	}

	for _, query := range equivalentQueries {
		queryMsg := &pgproto3.Query{String: query}
		queryBuf := make([]byte, 0, 1024)
		queryBuf, err = queryMsg.Encode(queryBuf)
		require.NoError(t, err, "Failed to encode equivalent query")

		_, err = conn.Write(queryBuf)
		require.NoError(t, err, "Failed to send equivalent query")

		time.Sleep(200 * time.Millisecond)
	}

	// Wait a bit more for processing
	time.Sleep(1 * time.Second)

	// Get updated normalization data
	finalNormalizedData := testLogger.GetNormalizedData()

	// The first query and the two equivalent queries should have the same hash
	if len(finalNormalizedData) >= 3 {
		firstQueryHash := finalNormalizedData[0].Hash.Value()
		equivalentHash1 := finalNormalizedData[len(finalNormalizedData)-2].Hash.Value()
		equivalentHash2 := finalNormalizedData[len(finalNormalizedData)-1].Hash.Value()

		assert.Equal(t, firstQueryHash, equivalentHash1, "Equivalent queries should have same hash")
		assert.Equal(t, firstQueryHash, equivalentHash2, "Equivalent queries should have same hash")

		t.Logf("✓ Hash consistency verified: %s", firstQueryHash)
	}

	t.Log("=== Query Normalization Integration Test Complete ===")
}
