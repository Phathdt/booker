package consumer

import (
	"context"
	"testing"

	"booker/pkg/logger"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

// MockLogger for testing
type MockLogger struct {
	messages []string
}

func (m *MockLogger) With(args ...interface{}) logger.Logger {
	return m
}

func (m *MockLogger) WithGroup(name string) logger.Logger {
	return m
}

func (m *MockLogger) Debug(msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

func (m *MockLogger) Info(msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

func (m *MockLogger) Warn(msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

func (m *MockLogger) Error(msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

func (m *MockLogger) Fatal(msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

func (m *MockLogger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

func (m *MockLogger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

func (m *MockLogger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

func (m *MockLogger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	m.messages = append(m.messages, msg)
}

// --- Test Cases ---

func TestNewNATSConsumer_CreatesConsumer(t *testing.T) {
	// Create minimal mocks - just pass nil to test basic initialization
	var js nats.JetStreamContext
	handler := &EventHandler{}
	log := &MockLogger{}

	consumer := NewNATSConsumer(js, handler, log)

	assert.NotNil(t, consumer)
	assert.Equal(t, handler, consumer.handler)
	assert.Nil(t, consumer.cancel)
	assert.Len(t, consumer.subs, 0)
}

func TestNATSConsumer_Stop_WithoutStart(t *testing.T) {
	var js nats.JetStreamContext
	handler := &EventHandler{}
	log := &MockLogger{}

	consumer := NewNATSConsumer(js, handler, log)

	// Should not panic when stopping without starting
	consumer.Stop()
	assert.Nil(t, consumer.cancel)
}
