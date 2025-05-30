package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockLogger implements Logger interface for testing
type mockLogger struct {
	messages []logEntry
	fields   map[string]interface{}
}

type logEntry struct {
	level   string
	message string
	args    []interface{}
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		messages: make([]logEntry, 0),
		fields:   make(map[string]interface{}),
	}
}

func (m *mockLogger) Info(msg string, args ...interface{}) {
	m.messages = append(m.messages, logEntry{
		level:   "INFO",
		message: msg,
		args:    args,
	})
}

func (m *mockLogger) Error(msg string, args ...interface{}) {
	m.messages = append(m.messages, logEntry{
		level:   "ERROR",
		message: msg,
		args:    args,
	})
}

func (m *mockLogger) Debug(msg string, args ...interface{}) {
	m.messages = append(m.messages, logEntry{
		level:   "DEBUG",
		message: msg,
		args:    args,
	})
}

func (m *mockLogger) WithField(key string, value interface{}) Logger {
	newFields := make(map[string]interface{})
	for k, v := range m.fields {
		newFields[k] = v
	}
	newFields[key] = value

	return &mockLogger{
		messages: m.messages,
		fields:   newFields,
	}
}

func (m *mockLogger) hasLogLevel(level string) bool {
	for _, entry := range m.messages {
		if entry.level == level {
			return true
		}
	}
	return false
}

func (m *mockLogger) hasMessage(message string) bool {
	for _, entry := range m.messages {
		if entry.message == message {
			return true
		}
	}
	return false
}

func (m *mockLogger) getFieldValue(key string) (interface{}, bool) {
	value, exists := m.fields[key]
	return value, exists
}

func TestNewSimpleLogger(t *testing.T) {
	logger := NewSimpleLogger()

	assert.NotNil(t, logger)
	assert.Implements(t, (*Logger)(nil), logger)
}

func TestLogger_Interface(t *testing.T) {
	mock := newMockLogger()

	// Test Info
	mock.Info("test info message")
	assert.True(t, mock.hasLogLevel("INFO"))
	assert.True(t, mock.hasMessage("test info message"))

	// Test Error
	mock.Error("test error message")
	assert.True(t, mock.hasLogLevel("ERROR"))
	assert.True(t, mock.hasMessage("test error message"))

	// Test Debug
	mock.Debug("test debug message")
	assert.True(t, mock.hasLogLevel("DEBUG"))
	assert.True(t, mock.hasMessage("test debug message"))
}

func TestLogger_WithField(t *testing.T) {
	mock := newMockLogger()

	// Add field and verify it's isolated
	loggerWithField := mock.WithField("key1", "value1")

	// Verify the original logger doesn't have the field
	_, exists := mock.getFieldValue("key1")
	assert.False(t, exists)

	// Verify the new logger has the field
	mockWithField := loggerWithField.(*mockLogger)
	value, exists := mockWithField.getFieldValue("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)
}

func TestLogger_MultipleFields(t *testing.T) {
	mock := newMockLogger()

	// Chain multiple fields
	loggerWithFields := mock.WithField("key1", "value1").
		WithField("key2", 42).
		WithField("key3", true)

	mockWithFields := loggerWithFields.(*mockLogger)

	// Verify all fields are present
	value1, exists1 := mockWithFields.getFieldValue("key1")
	assert.True(t, exists1)
	assert.Equal(t, "value1", value1)

	value2, exists2 := mockWithFields.getFieldValue("key2")
	assert.True(t, exists2)
	assert.Equal(t, 42, value2)

	value3, exists3 := mockWithFields.getFieldValue("key3")
	assert.True(t, exists3)
	assert.Equal(t, true, value3)
}

func TestLogger_FieldIsolation(t *testing.T) {
	baseMock := newMockLogger()

	// Create two loggers with different field values
	logger1 := baseMock.WithField("logger_id", "1")
	logger2 := baseMock.WithField("logger_id", "2")

	// Verify isolation
	mock1 := logger1.(*mockLogger)
	mock2 := logger2.(*mockLogger)

	value1, _ := mock1.getFieldValue("logger_id")
	value2, _ := mock2.getFieldValue("logger_id")

	assert.Equal(t, "1", value1)
	assert.Equal(t, "2", value2)

	// Verify base logger is unchanged
	_, exists := baseMock.getFieldValue("logger_id")
	assert.False(t, exists)
}
