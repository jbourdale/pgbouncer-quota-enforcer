package infrastructure

import (
	"encoding/hex"
	"fmt"
	"pgbouncer-quota-enforcer/internal/app/domain"
	"pgbouncer-quota-enforcer/pkg/logger"
)

// StandardByteLogger implements domain.ByteLogger
type StandardByteLogger struct {
	logger logger.Logger
}

// NewStandardByteLogger creates a new StandardByteLogger
func NewStandardByteLogger(log logger.Logger) domain.ByteLogger {
	return &StandardByteLogger{
		logger: log,
	}
}

// LogBytes logs the received bytes with connection information
func (l *StandardByteLogger) LogBytes(connectionID string, data []byte) error {
	if len(data) == 0 {
		return nil
	}

	// Create a logger with connection context
	connLogger := l.logger.WithField("connection_id", connectionID)

	// Log basic information
	connLogger.Info("Received bytes",
		"length", len(data),
		"hex_preview", l.formatHexPreview(data, 32),
		"ascii_preview", l.formatASCIIPreview(data, 32),
	)

	// For debugging purposes, log full hex dump for small packets
	if len(data) <= 256 {
		connLogger.Debug("Full packet hex dump",
			"hex_data", hex.EncodeToString(data),
		)
	}

	return nil
}

// formatHexPreview creates a hex preview of the data (limited to maxBytes)
func (l *StandardByteLogger) formatHexPreview(data []byte, maxBytes int) string {
	if len(data) == 0 {
		return ""
	}

	end := len(data)
	if end > maxBytes {
		end = maxBytes
	}

	preview := hex.EncodeToString(data[:end])
	if len(data) > maxBytes {
		preview += "..."
	}

	return preview
}

// formatASCIIPreview creates an ASCII preview of the data (limited to maxBytes)
func (l *StandardByteLogger) formatASCIIPreview(data []byte, maxBytes int) string {
	if len(data) == 0 {
		return `""`
	}

	end := len(data)
	if end > maxBytes {
		end = maxBytes
	}

	result := make([]byte, 0, end)
	for _, b := range data[:end] {
		if b >= 32 && b <= 126 { // Printable ASCII characters
			result = append(result, b)
		} else {
			result = append(result, '.')
		}
	}

	preview := string(result)
	if len(data) > maxBytes {
		preview += "..."
	}

	return fmt.Sprintf("%q", preview)
}
