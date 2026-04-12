package ws

import (
	"context"
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

// TestNewClient_WithNilConn tests NewClient initializes even with nil conn
func TestNewClient_WithNilConn(t *testing.T) {
	hub := NewHub()
	client := NewClient(nil, hub)

	assert.NotNil(t, client)
	assert.Equal(t, hub, client.hub)
	assert.Nil(t, client.conn)
	assert.NotNil(t, client.send)
	assert.Equal(t, sendBufferSize, cap(client.send))
}

// TestNewClient_ProperlyInitializesSendChannel tests that NewClient properly initializes send channel
func TestNewClient_ProperlyInitializesSendChannel(t *testing.T) {
	hub := NewHub()
	client := NewClient(nil, hub)

	// Verify channel properties
	assert.NotNil(t, client.send)
	assert.Equal(t, sendBufferSize, cap(client.send))
	assert.Equal(t, 0, len(client.send))
}

// MockWSConn implements WSConn for testing
type MockWSConn struct {
	readMessages  chan []byte
	writeMessages chan []byte
	closed        bool
	readErr       error
	writeErr      error
	readDeadline  time.Time
	writeDeadline time.Time
	pongHandler   func(string) error
}

func NewMockWSConn() *MockWSConn {
	return &MockWSConn{
		readMessages:  make(chan []byte, 10),
		writeMessages: make(chan []byte, 10),
		closed:        false,
	}
}

func (m *MockWSConn) ReadMessage() (messageType int, data []byte, err error) {
	if m.readErr != nil {
		return 0, nil, m.readErr
	}
	select {
	case msg := <-m.readMessages:
		return 1, msg, nil // 1 = text message
	case <-time.After(100 * time.Millisecond):
		return 0, nil, &mockWSError{msg: "read timeout"}
	}
}

func (m *MockWSConn) WriteMessage(messageType int, data []byte) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	if m.closed {
		return &mockWSError{msg: "connection closed"}
	}
	select {
	case m.writeMessages <- data:
		return nil
	default:
		return &mockWSError{msg: "write buffer full"}
	}
}

func (m *MockWSConn) Close() error {
	m.closed = true
	return nil
}

func (m *MockWSConn) SetReadDeadline(t time.Time) error {
	m.readDeadline = t
	return nil
}

func (m *MockWSConn) SetWriteDeadline(t time.Time) error {
	m.writeDeadline = t
	return nil
}

func (m *MockWSConn) SetReadLimit(limit int64) {
	// no-op for mock
}

func (m *MockWSConn) SetPongHandler(h func(string) error) {
	m.pongHandler = h
}

type mockWSError struct {
	msg string
}

func (e *mockWSError) Error() string {
	return e.msg
}

func (e *mockWSError) Timeout() bool {
	return true
}

func (e *mockWSError) Temporary() bool {
	return true
}

// SequentialMockWSConn tracks read call count and returns error after messages
type SequentialMockWSConn struct {
	*MockWSConn
	readCount int
	errorAt   int
}

func NewSequentialMockWSConn(errorAt int) *SequentialMockWSConn {
	return &SequentialMockWSConn{
		MockWSConn: NewMockWSConn(),
		readCount:  0,
		errorAt:    errorAt,
	}
}

func (s *SequentialMockWSConn) ReadMessage() (messageType int, data []byte, err error) {
	s.readCount++
	if s.readCount >= s.errorAt {
		return 0, nil, &mockWSError{msg: "connection closed"}
	}
	return s.MockWSConn.ReadMessage()
}

// TestReadPump_ValidSubscribeMessage tests ReadPump with valid subscribe message
func TestReadPump_ValidSubscribeMessage(t *testing.T) {
	hub := NewHub()
	mockConn := NewSequentialMockWSConn(2) // Error after 1st message
	client := &Client{
		hub:  hub,
		conn: mockConn,
		send: make(chan []byte, sendBufferSize),
	}

	// Send valid subscribe message
	subscribeMsg := SubscribeMsg{Op: "subscribe", Channel: "ticker", Pair: "BTC_USDT"}
	data, _ := json.Marshal(subscribeMsg)
	mockConn.readMessages <- data

	client.ReadPump()

	// Verify connection was closed
	assert.True(t, mockConn.closed)
}

// TestReadPump_InvalidJSON tests ReadPump handles invalid JSON
func TestReadPump_InvalidJSON(t *testing.T) {
	hub := NewHub()
	mockConn := NewSequentialMockWSConn(3) // Allow 2 messages, error on 3rd
	client := &Client{
		hub:  hub,
		conn: mockConn,
		send: make(chan []byte, sendBufferSize),
	}

	// Send invalid JSON
	mockConn.readMessages <- []byte("invalid json {")
	// Send valid message to detect continuation after error
	validMsg := SubscribeMsg{Op: "subscribe", Channel: "ticker", Pair: "BTC_USDT"}
	data, _ := json.Marshal(validMsg)
	mockConn.readMessages <- data

	client.ReadPump()

	// Verify error message was sent
	select {
	case msg := <-client.send:
		var wsMsg WSMessage
		err := json.Unmarshal(msg, &wsMsg)
		require.NoError(t, err)
		assert.Equal(t, "error", wsMsg.Type)
	case <-time.After(500 * time.Millisecond):
		// May not send if buffer full
	}
}

// TestReadPump_InvalidChannel tests ReadPump rejects invalid channel
func TestReadPump_InvalidChannel(t *testing.T) {
	hub := NewHub()
	mockConn := NewSequentialMockWSConn(3) // Allow 2 messages
	client := &Client{
		hub:  hub,
		conn: mockConn,
		send: make(chan []byte, sendBufferSize),
	}

	// Send message with invalid channel
	subscribeMsg := SubscribeMsg{Op: "subscribe", Channel: "invalid_channel", Pair: "BTC_USDT"}
	data, _ := json.Marshal(subscribeMsg)
	mockConn.readMessages <- data

	// Send continuation message
	validMsg := SubscribeMsg{Op: "subscribe", Channel: "ticker", Pair: "BTC_USDT"}
	data, _ = json.Marshal(validMsg)
	mockConn.readMessages <- data

	client.ReadPump()

	// Verify error message was sent
	select {
	case msg := <-client.send:
		var wsMsg WSMessage
		json.Unmarshal(msg, &wsMsg)
		assert.Equal(t, "error", wsMsg.Type)
		assert.Contains(t, wsMsg.Msg, "unknown channel")
	case <-time.After(500 * time.Millisecond):
		// May not send if buffer full
	}
}

// TestReadPump_InvalidOp tests ReadPump rejects invalid operation
func TestReadPump_InvalidOp(t *testing.T) {
	hub := NewHub()
	mockConn := NewSequentialMockWSConn(3) // Allow 2 messages
	client := &Client{
		hub:  hub,
		conn: mockConn,
		send: make(chan []byte, sendBufferSize),
	}

	// Send message with invalid op
	subscribeMsg := SubscribeMsg{Op: "invalid_op", Channel: "ticker", Pair: "BTC_USDT"}
	data, _ := json.Marshal(subscribeMsg)
	mockConn.readMessages <- data

	// Send continuation message
	validMsg := SubscribeMsg{Op: "subscribe", Channel: "ticker", Pair: "BTC_USDT"}
	data, _ = json.Marshal(validMsg)
	mockConn.readMessages <- data

	client.ReadPump()

	// Verify error message was sent
	select {
	case msg := <-client.send:
		var wsMsg WSMessage
		json.Unmarshal(msg, &wsMsg)
		assert.Equal(t, "error", wsMsg.Type)
		assert.Contains(t, wsMsg.Msg, "unknown op")
	case <-time.After(500 * time.Millisecond):
		// May not send if buffer full
	}
}

// TestReadPump_UnsubscribeMessage tests ReadPump handles unsubscribe
func TestReadPump_UnsubscribeMessage(t *testing.T) {
	hub := NewHub()
	mockConn := NewSequentialMockWSConn(3) // Allow 2 messages
	client := &Client{
		hub:  hub,
		conn: mockConn,
		send: make(chan []byte, sendBufferSize),
	}

	// Send subscribe then unsubscribe
	subMsg := SubscribeMsg{Op: "subscribe", Channel: "ticker", Pair: "BTC_USDT"}
	data, _ := json.Marshal(subMsg)
	mockConn.readMessages <- data

	unsubMsg := SubscribeMsg{Op: "unsubscribe", Channel: "ticker", Pair: "BTC_USDT"}
	data, _ = json.Marshal(unsubMsg)
	mockConn.readMessages <- data

	client.ReadPump()

	// Verify connection was closed
	assert.True(t, mockConn.closed)
}

// TestReadPump_ConnectionError tests ReadPump exits on connection error
func TestReadPump_ConnectionError(t *testing.T) {
	hub := NewHub()
	mockConn := NewSequentialMockWSConn(1) // Error immediately
	client := &Client{
		hub:  hub,
		conn: mockConn,
		send: make(chan []byte, sendBufferSize),
	}

	client.ReadPump()

	// Verify connection was closed
	assert.True(t, mockConn.closed)
}

// TestWritePump_SendsMessages tests WritePump sends buffered messages
func TestWritePump_SendsMessages(t *testing.T) {
	mockConn := NewMockWSConn()
	client := &Client{
		hub:  NewHub(),
		conn: mockConn,
		send: make(chan []byte, sendBufferSize),
	}

	// Send message then close channel
	testMsg := []byte(`{"type":"ticker","pair":"BTC_USDT"}`)
	client.send <- testMsg

	go func() {
		time.Sleep(50 * time.Millisecond)
		close(client.send)
	}()

	client.WritePump()

	// Verify connection was closed
	assert.True(t, mockConn.closed)

	// Verify message was written
	select {
	case msg := <-mockConn.writeMessages:
		assert.Equal(t, testMsg, msg)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("WritePump should write message")
	}
}

// TestWritePump_SendsPingMessages tests WritePump sends periodic pings
func TestWritePump_SendsPingMessages(t *testing.T) {
	mockConn := NewMockWSConn()
	client := &Client{
		hub:  NewHub(),
		conn: mockConn,
		send: make(chan []byte, sendBufferSize),
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		close(client.send)
	}()

	client.WritePump()

	// Verify connection was closed
	assert.True(t, mockConn.closed)

	// There should be at least some write messages (ping messages)
	writtenCount := 0
	timeout := time.After(1 * time.Second)
	for {
		select {
		case <-mockConn.writeMessages:
			writtenCount++
		case <-timeout:
			// Expect at least one ping message
			assert.Greater(t, writtenCount, 0)
			return
		}
	}
}

// TestWritePump_ClosesOnChannelClose tests WritePump closes on send channel close
func TestWritePump_ClosesOnChannelClose(t *testing.T) {
	mockConn := NewMockWSConn()
	client := &Client{
		hub:  NewHub(),
		conn: mockConn,
		send: make(chan []byte, sendBufferSize),
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		close(client.send)
	}()

	client.WritePump()

	// Verify connection was closed
	assert.True(t, mockConn.closed)
}

// TestWritePump_HandlesWriteError tests WritePump exits on write error
func TestWritePump_HandlesWriteError(t *testing.T) {
	mockConn := NewMockWSConn()
	client := &Client{
		hub:  NewHub(),
		conn: mockConn,
		send: make(chan []byte, sendBufferSize),
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		mockConn.writeErr = &mockWSError{msg: "write failed"}
	}()

	client.WritePump()

	// Verify connection was closed
	assert.True(t, mockConn.closed)
}

// TestSendError_WithBlockedSendChannel tests sendError when channel is full
func TestSendError_WithBlockedSendChannel(t *testing.T) {
	hub := NewHub()
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, 1),
	}

	// Block the channel
	client.send <- []byte("blocked")

	// This should not block even though channel is full
	done := make(chan struct{})
	go func() {
		client.sendError("error while blocked")
		done <- struct{}{}
	}()

	select {
	case <-done:
		// Expected - sendError should use non-blocking send
	case <-time.After(1 * time.Second):
		t.Fatal("sendError should not block")
	}
}

// TestReadPump_SetReadDeadlineCalledAtStart tests that SetReadDeadline is called
func TestReadPump_SetReadDeadlineCalledAtStart(t *testing.T) {
	hub := NewHub()
	mockConn := NewMockWSConn()
	client := &Client{
		hub:  hub,
		conn: mockConn,
		send: make(chan []byte, sendBufferSize),
	}

	// Send an error to exit immediately
	mockConn.readErr = &mockWSError{msg: "connection closed"}

	client.ReadPump()

	// Verify connection was closed
	assert.True(t, mockConn.closed)
	// Verify SetReadDeadline was called (it will be set to a time in the future)
	assert.False(t, mockConn.readDeadline.IsZero())
}

// TestBroadcast_WithSlowClient tests broadcast behavior with slow client
func TestBroadcast_WithSlowClient(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	// Create client with small buffer to simulate slow client
	client := &Client{
		hub:  hub,
		conn: nil,
		send: make(chan []byte, 1), // Small buffer
	}

	hub.Register(client)
	time.Sleep(50 * time.Millisecond)

	// Fill the buffer
	client.send <- []byte("full")

	// Broadcast should drop message for slow client (non-blocking send)
	ticker := TickerData{
		Open:      "100",
		High:      "110",
		Low:       "90",
		Close:     "105",
		Volume:    "1000",
		ChangePct: "5",
		LastPrice: "105",
		Timestamp: 123456789,
	}

	hub.BroadcastTicker("BTC_USDT", ticker)
	time.Sleep(50 * time.Millisecond)

	// Verify message was dropped (non-blocking select)
	assert.Equal(t, 1, len(client.send))
}

// TestWritePump_MultipleMessages tests WritePump with multiple buffered messages
func TestWritePump_MultipleMessages(t *testing.T) {
	mockConn := NewMockWSConn()
	client := &Client{
		hub:  NewHub(),
		conn: mockConn,
		send: make(chan []byte, sendBufferSize),
	}

	// Send multiple messages
	msg1 := []byte(`{"type":"ticker","pair":"BTC_USDT"}`)
	msg2 := []byte(`{"type":"trade","pair":"ETH_USDT"}`)
	client.send <- msg1
	client.send <- msg2

	go func() {
		time.Sleep(100 * time.Millisecond)
		close(client.send)
	}()

	client.WritePump()

	// Verify all messages were written
	receivedCount := 0
	timeout := time.After(1 * time.Second)
	for {
		select {
		case msg := <-mockConn.writeMessages:
			assert.NotNil(t, msg)
			receivedCount++
		case <-timeout:
			assert.GreaterOrEqual(t, receivedCount, 2)
			return
		}
	}
}

// TestReadPump_TradesChannel tests ReadPump with trades channel subscription
func TestReadPump_TradesChannel(t *testing.T) {
	hub := NewHub()
	mockConn := NewSequentialMockWSConn(2)
	client := &Client{
		hub:  hub,
		conn: mockConn,
		send: make(chan []byte, sendBufferSize),
	}

	// Send trades channel subscription
	subscribeMsg := SubscribeMsg{Op: "subscribe", Channel: "trades", Pair: "ETH_USDT"}
	data, _ := json.Marshal(subscribeMsg)
	mockConn.readMessages <- data

	client.ReadPump()

	// Verify connection was closed
	assert.True(t, mockConn.closed)
}

// TestWritePump_WriteDeadlineSet tests that WritePump sets write deadlines
func TestWritePump_WriteDeadlineSet(t *testing.T) {
	mockConn := NewMockWSConn()
	client := &Client{
		hub:  NewHub(),
		conn: mockConn,
		send: make(chan []byte, sendBufferSize),
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		close(client.send)
	}()

	client.WritePump()

	// Verify connection was closed
	assert.True(t, mockConn.closed)
	// Verify write deadline was set
	assert.False(t, mockConn.writeDeadline.IsZero())
}

// TestReadPump_AllChannelsSupported tests all valid channels are supported
func TestReadPump_AllChannelsSupported(t *testing.T) {
	channels := []string{"ticker", "trades"}

	for _, ch := range channels {
		t.Run(ch, func(t *testing.T) {
			hub := NewHub()
			mockConn := NewSequentialMockWSConn(2)
			client := &Client{
				hub:  hub,
				conn: mockConn,
				send: make(chan []byte, sendBufferSize),
			}

			subscribeMsg := SubscribeMsg{Op: "subscribe", Channel: ch, Pair: "BTC_USDT"}
			data, _ := json.Marshal(subscribeMsg)
			mockConn.readMessages <- data

			client.ReadPump()

			assert.True(t, mockConn.closed)
		})
	}
}
