package adapters

import (
	"pgbouncer-quota-enforcer/internal/app/domain"
	"pgbouncer-quota-enforcer/pkg/logger"
	"strings"
)

// StandardQueryLogger implements domain.QueryLogger
type StandardQueryLogger struct {
	logger     logger.Logger
	normalizer domain.QueryNormalizer
}

// NewStandardQueryLogger creates a new StandardQueryLogger
func NewStandardQueryLogger(log logger.Logger, normalizer domain.QueryNormalizer) domain.QueryLogger {
	return &StandardQueryLogger{
		logger:     log,
		normalizer: normalizer,
	}
}

// LogQuery logs a SQL query with connection information
func (l *StandardQueryLogger) LogQuery(connectionID string, query string) error {
	if query == "" {
		return nil
	}

	// Create a logger with connection context
	connLogger := l.logger.WithField("connection_id", connectionID)

	// Clean up the query for logging (remove extra whitespace, newlines)
	cleanQuery := strings.TrimSpace(strings.ReplaceAll(query, "\n", " "))

	// Truncate very long queries for readability
	const maxQueryLength = 500
	if len(cleanQuery) > maxQueryLength {
		cleanQuery = cleanQuery[:maxQueryLength] + "..."
	}

	connLogger.Info("SQL Query received",
		"query", cleanQuery,
		"query_length", len(query),
	)

	return nil
}

// LogNormalizedQuery logs a normalized SQL query with hash
func (l *StandardQueryLogger) LogNormalizedQuery(connectionID string, normalizedQuery domain.NormalizedQuery) error {
	// Create a logger with connection context
	connLogger := l.logger.WithField("connection_id", connectionID)

	// Log the normalized query with hash
	connLogger.Info("Normalized SQL Query",
		"original_query", normalizedQuery.Original,
		"normalized_query", normalizedQuery.Normalized,
		"query_hash", normalizedQuery.Hash.Value(),
	)

	return nil
}

// LogProtocolMessage logs other protocol messages (startup, auth, etc.)
func (l *StandardQueryLogger) LogProtocolMessage(connectionID string, messageType string, details map[string]interface{}) error {
	// Create a logger with connection context
	connLogger := l.logger.WithField("connection_id", connectionID)

	// Convert details to a more readable format
	logFields := make([]interface{}, 0, len(details)*2+2)
	logFields = append(logFields, "message_type", messageType)

	for key, value := range details {
		logFields = append(logFields, key, value)
	}

	connLogger.Info("PostgreSQL protocol message", logFields...)

	return nil
}
