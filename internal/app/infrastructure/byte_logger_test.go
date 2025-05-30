package infrastructure

import (
	"pgbouncer-quota-enforcer/pkg/logger"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockLogger implements logger.Logger for testing
type mockLogger struct {
	messages *[]string // Use pointer to share messages across WithField calls
	fields   map[string]interface{}
}

func newMockLogger() *mockLogger {
	messages := make([]string, 0)
	return &mockLogger{
		messages: &messages,
		fields:   make(map[string]interface{}),
	}
}

func (m *mockLogger) Info(msg string, args ...interface{}) {
	*m.messages = append(*m.messages, "INFO: "+msg)
}

func (m *mockLogger) Error(msg string, args ...interface{}) {
	*m.messages = append(*m.messages, "ERROR: "+msg)
}

func (m *mockLogger) Debug(msg string, args ...interface{}) {
	*m.messages = append(*m.messages, "DEBUG: "+msg)
}

func (m *mockLogger) WithField(key string, value interface{}) logger.Logger {
	newFields := make(map[string]interface{})
	for k, v := range m.fields {
		newFields[k] = v
	}
	newFields[key] = value

	return &mockLogger{
		messages: m.messages, // Share the same messages slice
		fields:   newFields,
	}
}

func (m *mockLogger) hasMessage(substring string) bool {
	for _, msg := range *m.messages {
		if containsString(msg, substring) {
			return true
		}
	}
	return false
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(len(substr) == 0 || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestStandardByteLogger_LogBytes(t *testing.T) {
	tests := []struct {
		name         string
		connectionID string
		data         []byte
		expectLogged bool
	}{
		{
			name:         "log simple data",
			connectionID: "conn_1",
			data:         []byte("Hello World"),
			expectLogged: true,
		},
		{
			name:         "log binary data",
			connectionID: "conn_2",
			data:         []byte{0x00, 0x01, 0x02, 0xFF},
			expectLogged: true,
		},
		{
			name:         "empty data should not log",
			connectionID: "conn_3",
			data:         []byte{},
			expectLogged: false,
		},
		{
			name:         "nil data should not log",
			connectionID: "conn_4",
			data:         nil,
			expectLogged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLog := newMockLogger()
			byteLogger := NewStandardByteLogger(mockLog)

			err := byteLogger.LogBytes(tt.connectionID, tt.data)
			assert.NoError(t, err)

			if tt.expectLogged {
				assert.True(t, mockLog.hasMessage("Received bytes"))
			} else {
				assert.False(t, mockLog.hasMessage("Received bytes"))
			}
		})
	}
}

func TestStandardByteLogger_HexPreview(t *testing.T) {
	mockLog := newMockLogger()
	byteLogger := NewStandardByteLogger(mockLog).(*StandardByteLogger)

	tests := []struct {
		name     string
		data     []byte
		maxBytes int
		expected string
	}{
		{
			name:     "simple hex",
			data:     []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}, // "Hello"
			maxBytes: 10,
			expected: "48656c6c6f",
		},
		{
			name:     "truncated hex",
			data:     []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}, // "Hello"
			maxBytes: 2,
			expected: "4865...",
		},
		{
			name:     "empty data",
			data:     []byte{},
			maxBytes: 10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := byteLogger.formatHexPreview(tt.data, tt.maxBytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStandardByteLogger_ASCIIPreview(t *testing.T) {
	mockLog := newMockLogger()
	byteLogger := NewStandardByteLogger(mockLog).(*StandardByteLogger)

	tests := []struct {
		name     string
		data     []byte
		maxBytes int
		expected string
	}{
		{
			name:     "printable ASCII",
			data:     []byte("Hello"),
			maxBytes: 10,
			expected: `"Hello"`,
		},
		{
			name:     "mixed data",
			data:     []byte{0x48, 0x65, 0x00, 0x6c, 0x6f}, // "He\x00lo"
			maxBytes: 10,
			expected: `"He.lo"`,
		},
		{
			name:     "truncated ASCII",
			data:     []byte("Hello World"),
			maxBytes: 5,
			expected: `"Hello..."`,
		},
		{
			name:     "empty data",
			data:     []byte{},
			maxBytes: 10,
			expected: `""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := byteLogger.formatASCIIPreview(tt.data, tt.maxBytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStandardByteLogger_LargePacket(t *testing.T) {
	mockLog := newMockLogger()
	byteLogger := NewStandardByteLogger(mockLog)

	// Create a large packet (>256 bytes)
	largeData := make([]byte, 300)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	err := byteLogger.LogBytes("conn_large", largeData)
	assert.NoError(t, err)

	// Should log basic info
	assert.True(t, mockLog.hasMessage("Received bytes"))
}

func TestStandardByteLogger_SmallPacket(t *testing.T) {
	mockLog := newMockLogger()
	byteLogger := NewStandardByteLogger(mockLog)

	// Create a small packet (<=256 bytes)
	smallData := []byte("small packet")

	err := byteLogger.LogBytes("conn_small", smallData)
	assert.NoError(t, err)

	// Should log basic info
	assert.True(t, mockLog.hasMessage("Received bytes"))
}
