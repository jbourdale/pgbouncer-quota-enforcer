package infrastructure

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

// LoggingConnectionHandler implements domain.ConnectionHandler
type LoggingConnectionHandler struct {
	byteLogger   domain.ByteLogger
	logger       logger.Logger
	bufferSize   int
	readTimeout  time.Duration
	connectionID int64 // Atomic counter for connection IDs
}

// NewLoggingConnectionHandler creates a new LoggingConnectionHandler
func NewLoggingConnectionHandler(byteLogger domain.ByteLogger, log logger.Logger) domain.ConnectionHandler {
	return &LoggingConnectionHandler{
		byteLogger:  byteLogger,
		logger:      log,
		bufferSize:  4096, // 4KB buffer for reading
		readTimeout: 30 * time.Second,
	}
}

// HandleConnection processes an incoming TCP connection
func (h *LoggingConnectionHandler) HandleConnection(ctx context.Context, conn net.Conn) error {
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

	connLogger.Info("New connection established")

	// Read data in a loop until connection is closed or context is cancelled
	buffer := make([]byte, h.bufferSize)

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

			// Read data from connection
			n, err := conn.Read(buffer)
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

				connLogger.Error("Error reading from connection: %v", err)
				return fmt.Errorf("error reading from connection: %w", err)
			}

			if n > 0 {
				// Log the received bytes
				if err := h.byteLogger.LogBytes(connectionID, buffer[:n]); err != nil {
					connLogger.Error("Error logging bytes: %v", err)
					// Continue processing even if logging fails
				}
			}
		}
	}
}
