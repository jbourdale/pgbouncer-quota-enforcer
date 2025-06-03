package adapters

import (
	"context"
	"fmt"
	"io"
	"net"
	"pgbouncer-quota-enforcer/internal/app/domain"
	"pgbouncer-quota-enforcer/pkg/logger"
	"sync/atomic"
	"time"
)

// PostgreSQLConnectionHandler implements domain.ConnectionHandler for PostgreSQL protocol
type PostgreSQLConnectionHandler struct {
	queryLogger  domain.QueryLogger
	normalizer   domain.QueryNormalizer
	logger       logger.Logger
	readTimeout  time.Duration
	connectionID int64 // Atomic counter for connection IDs
}

// NewPostgreSQLConnectionHandler creates a new PostgreSQL connection handler
func NewPostgreSQLConnectionHandler(queryLogger domain.QueryLogger, normalizer domain.QueryNormalizer, log logger.Logger) domain.ConnectionHandler {
	return &PostgreSQLConnectionHandler{
		queryLogger: queryLogger,
		normalizer:  normalizer,
		logger:      log,
		readTimeout: 30 * time.Second,
	}
}

// HandleConnection processes an incoming PostgreSQL connection
func (h *PostgreSQLConnectionHandler) HandleConnection(ctx context.Context, conn net.Conn) error {
	// Generate unique connection ID
	connID := atomic.AddInt64(&h.connectionID, 1)
	connectionID := fmt.Sprintf("conn_%d", connID)

	// Create logger with connection context
	connLogger := h.logger.WithField("connection_id", connectionID).
		WithField("remote_addr", conn.RemoteAddr().String())

	// Ensure connection is closed when done
	defer func() {
		if err := conn.Close(); err != nil {
			connLogger.Error("Error closing connection: %v", err)
		}
		connLogger.Info("Connection closed")
	}()

	connLogger.Info("New PostgreSQL connection established")

	// Create PostgreSQL protocol parser
	// Note: We're creating a dummy writer since we're only parsing, not responding
	parser := NewPostgreSQLParser(conn, io.Discard)

	// Process messages in a loop until connection is closed or context is cancelled
	for {
		select {
		case <-ctx.Done():
			connLogger.Info("Connection handler stopped due to context cancellation")
			return ctx.Err()
		default:
			// Set read timeout
			if err := conn.SetReadDeadline(time.Now().Add(h.readTimeout)); err != nil {
				connLogger.Error("Failed to set read deadline: %v", err)
				return fmt.Errorf("failed to set read deadline: %w", err)
			}

			// Read and parse PostgreSQL message
			message, err := parser.ReadMessage()
			if err != nil {
				if err == io.EOF {
					connLogger.Info("Connection closed by client")
					return nil
				}

				// Check if it's a timeout error (expected during graceful shutdown)
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// Continue loop to check context cancellation
					continue
				}

				connLogger.Error("Error parsing PostgreSQL message: %v", err)
				return fmt.Errorf("error parsing PostgreSQL message: %w", err)
			}

			// Process the parsed message
			if err := h.processMessage(connectionID, message); err != nil {
				connLogger.Error("Error processing message: %v", err)
				// Continue processing even if logging fails
			}
		}
	}
}

// processMessage handles different types of PostgreSQL messages
func (h *PostgreSQLConnectionHandler) processMessage(connectionID string, message *ParsedMessage) error {
	switch message.Type {
	case "Query", "Parse":
		// Log and normalize SQL queries
		if message.Query != "" {
			// Log the original query
			if err := h.queryLogger.LogQuery(connectionID, message.Query); err != nil {
				h.logger.Error("Failed to log query: %v", err)
			}

			// Normalize the query and log normalized version
			normalizedQuery, err := h.normalizer.Normalize(message.Query)
			if err != nil {
				h.logger.Error("Failed to normalize query: %v", err)
				// Continue processing even if normalization fails
			} else {
				if err := h.queryLogger.LogNormalizedQuery(connectionID, normalizedQuery); err != nil {
					h.logger.Error("Failed to log normalized query: %v", err)
				}
			}
		}
	default:
		// Log other protocol messages
		return h.queryLogger.LogProtocolMessage(connectionID, message.Type, message.Details)
	}

	return nil
}
