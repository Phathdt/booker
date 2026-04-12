package ws

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewClient_InitializesSendChannel tests that NewClient initializes the send channel
func TestNewClient_InitializesSendChannel(t *testing.T) {
	hub := NewHub()

	// Create a minimal client - we can't use websocket.Conn directly without full setup
	// So we'll test the internal structure
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, sendBufferSize),
	}

	assert.NotNil(t, client.send)
	assert.Equal(t, sendBufferSize, cap(client.send))
	assert.Equal(t, hub, client.hub)
}

// TestClient_SendError_WritesToChannel tests that sendError writes to send channel
func TestClient_SendError_WritesToChannel(t *testing.T) {
	hub := NewHub()
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, sendBufferSize),
	}

	client.sendError("test error")

	select {
	case data := <-client.send:
		var msg WSMessage
		err := json.Unmarshal(data, &msg)
		require.NoError(t, err)
		assert.Equal(t, "error", msg.Type)
		assert.Equal(t, "test error", msg.Msg)
	case <-time.After(1 * time.Second):
		t.Fatal("sendError should send message")
	}
}

// TestClient_SendError_NonBlockingWhenFullChannel tests non-blocking behavior when channel full
func TestClient_SendError_NonBlockingWhenFullChannel(t *testing.T) {
	hub := NewHub()
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, 1),
	}

	// Fill the channel
	client.send <- []byte("full")

	// Should not block
	done := make(chan struct{})
	go func() {
		client.sendError("error when full")
		done <- struct{}{}
	}()

	select {
	case <-done:
		// Expected - should not block
	case <-time.After(1 * time.Second):
		t.Fatal("sendError should not block")
	}
}

// TestClient_SendError_WithSpecialCharacters tests error message with special chars
func TestClient_SendError_WithSpecialCharacters(t *testing.T) {
	hub := NewHub()
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, sendBufferSize),
	}

	specialMsg := `Error: "quote" 'apostrophe' \n newline \t tab`
	client.sendError(specialMsg)

	select {
	case data := <-client.send:
		var msg WSMessage
		err := json.Unmarshal(data, &msg)
		require.NoError(t, err)
		assert.Equal(t, "error", msg.Type)
		assert.Equal(t, specialMsg, msg.Msg)
	case <-time.After(1 * time.Second):
		t.Fatal("should handle special characters")
	}
}

// TestClient_SendError_WithUnicodeCharacters tests error message with unicode
func TestClient_SendError_WithUnicodeCharacters(t *testing.T) {
	hub := NewHub()
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, sendBufferSize),
	}

	unicodeMsg := "Error 错误 🚨 Ошибка"
	client.sendError(unicodeMsg)

	select {
	case data := <-client.send:
		var msg WSMessage
		err := json.Unmarshal(data, &msg)
		require.NoError(t, err)
		assert.Equal(t, "error", msg.Type)
		assert.Equal(t, unicodeMsg, msg.Msg)
	case <-time.After(1 * time.Second):
		t.Fatal("should handle unicode")
	}
}

// TestClient_SendError_EmptyMessage tests sending empty error message
func TestClient_SendError_EmptyMessage(t *testing.T) {
	hub := NewHub()
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, sendBufferSize),
	}

	client.sendError("")

	select {
	case data := <-client.send:
		var msg WSMessage
		err := json.Unmarshal(data, &msg)
		require.NoError(t, err)
		assert.Equal(t, "error", msg.Type)
		assert.Equal(t, "", msg.Msg)
	case <-time.After(1 * time.Second):
		t.Fatal("should send empty message")
	}
}

// TestClient_SendError_LongMessage tests sending a long error message
func TestClient_SendError_LongMessage(t *testing.T) {
	hub := NewHub()
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, sendBufferSize),
	}

	longMsg := string(make([]byte, 1000))
	client.sendError(longMsg)

	select {
	case data := <-client.send:
		var msg WSMessage
		err := json.Unmarshal(data, &msg)
		require.NoError(t, err)
		assert.Equal(t, "error", msg.Type)
		assert.Equal(t, longMsg, msg.Msg)
	case <-time.After(1 * time.Second):
		t.Fatal("should send long message")
	}
}

// TestClient_SendError_MultipleConcurrentSends tests concurrent sends
func TestClient_SendError_MultipleConcurrentSends(t *testing.T) {
	hub := NewHub()
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, 50),
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.sendError("concurrent error")
		}()
	}
	wg.Wait()

	count := 0
	timeout := time.After(1 * time.Second)
	for count < 10 {
		select {
		case <-client.send:
			count++
		case <-timeout:
			t.Fatal("timeout waiting for messages")
		}
	}

	assert.Equal(t, 10, count)
}

// TestClient_SendError_MessageFormat tests error message JSON format
func TestClient_SendError_MessageFormat(t *testing.T) {
	hub := NewHub()
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, sendBufferSize),
	}

	client.sendError("validation error")

	select {
	case data := <-client.send:
		var msg WSMessage
		err := json.Unmarshal(data, &msg)
		require.NoError(t, err)

		// Check all expected fields
		assert.Equal(t, "error", msg.Type)
		assert.Equal(t, "validation error", msg.Msg)
		assert.Equal(t, "", msg.Pair)
		assert.Nil(t, msg.Data)
	case <-time.After(1 * time.Second):
		t.Fatal("should send properly formatted message")
	}
}

// TestClient_SendError_DoesNotPanic tests sendError doesn't panic with edge cases
func TestClient_SendError_DoesNotPanic(t *testing.T) {
	hub := NewHub()
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, sendBufferSize),
	}

	// Should not panic with various inputs
	assert.NotPanics(t, func() {
		client.sendError("normal error")
	})

	assert.NotPanics(t, func() {
		client.sendError("")
	})

	assert.NotPanics(t, func() {
		client.sendError("error with\nnewlines\r\nand\ttabs")
	})
}

// TestClient_SendError_RapidSuccession tests rapid consecutive sends
func TestClient_SendError_RapidSuccession(t *testing.T) {
	hub := NewHub()
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, 20),
	}

	messages := []string{"error1", "error2", "error3", "error4", "error5"}

	for _, msg := range messages {
		client.sendError(msg)
	}

	time.Sleep(100 * time.Millisecond)

	for i, expectedMsg := range messages {
		select {
		case data := <-client.send:
			var msg WSMessage
			err := json.Unmarshal(data, &msg)
			require.NoError(t, err)
			assert.Equal(t, expectedMsg, msg.Msg)
		case <-time.After(1 * time.Second):
			t.Fatalf("failed to receive message %d", i)
		}
	}
}

// TestClient_Constants_Are_Defined tests that all constants are properly defined
func TestClient_Constants_Are_Defined(t *testing.T) {
	assert.Equal(t, time.Duration(10*time.Second), writeWait)
	assert.Equal(t, time.Duration(60*time.Second), pongWait)
	assert.Equal(t, time.Duration(30*time.Second), pingPeriod)
	assert.Equal(t, 256, sendBufferSize)
}

// TestClient_MultipleClients_Independent tests multiple clients are independent
func TestClient_MultipleClients_Independent(t *testing.T) {
	hub1 := NewHub()
	hub2 := NewHub()

	client1 := &Client{
		hub:  hub1,
		conn: nil,
		send: make(chan []byte, sendBufferSize),
	}

	client2 := &Client{
		hub:  hub2,
		conn: nil,
		send: make(chan []byte, sendBufferSize),
	}

	client1.sendError("error for client1")
	client2.sendError("error for client2")

	// Verify independence
	select {
	case data := <-client1.send:
		var msg WSMessage
		json.Unmarshal(data, &msg)
		assert.Equal(t, "error for client1", msg.Msg)
	case <-time.After(1 * time.Second):
		t.Fatal("client1 should receive its message")
	}

	select {
	case data := <-client2.send:
		var msg WSMessage
		json.Unmarshal(data, &msg)
		assert.Equal(t, "error for client2", msg.Msg)
	case <-time.After(1 * time.Second):
		t.Fatal("client2 should receive its message")
	}
}
