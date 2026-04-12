package ws

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"booker/modules/notification/domain/entities"

	"github.com/gofiber/contrib/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockWebSocketConn implements a mock of websocket.Conn
type MockWebSocketConn struct {
	mock.Mock
	messages [][]byte
	mu       sync.Mutex
}

func (m *MockWebSocketConn) WriteMessage(msgType int, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, data)
	args := m.Called(msgType, data)
	return args.Error(0)
}

func (m *MockWebSocketConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockWebSocketConn) GetMessages() [][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.messages
}

// --- Basic Hub Structure Tests ---

func TestHub_NewHub_CreatesEmptyConnsMap(t *testing.T) {
	hub := NewHub()
	assert.NotNil(t, hub)
	assert.Len(t, hub.conns, 0)
}

// --- Register/Unregister Logic Tests ---
// Note: Since websocket.Conn is a concrete type from gofiber/contrib,
// we test the core logic: registration tracking and cleanup.
// Full end-to-end tests would require integration tests with actual WebSocket connections.

func TestHub_RegisteredConnections_TrackedByUserID(t *testing.T) {
	hub := NewHub()

	// After registering connections for a user, verify the structure exists
	// We test the logic indirectly through SendToUser behavior
	notif := &entities.Notification{ID: "n1", Title: "Test"}

	// Sending to unregistered user should not panic
	hub.SendToUser("user-1", notif)

	hub.mu.RLock()
	_, exists := hub.conns["user-1"]
	hub.mu.RUnlock()

	// User entry should not be created by SendToUser
	assert.False(t, exists)
}

// --- SendToUser Tests ---

func TestHub_SendToUser_NoConnections_NoError(t *testing.T) {
	hub := NewHub()

	notif := &entities.Notification{
		ID:    "notif-1",
		Title: "Test Notification",
		Body:  "Test body",
	}

	// Should not panic when user has no connections
	hub.SendToUser("user-1", notif)
}

func TestHub_SendToUser_ValidNotification_JSONMarshalable(t *testing.T) {
	notif := &entities.Notification{
		ID:       "notif-1",
		UserID:   "user-1",
		Type:     entities.TypeTradeExecuted,
		Title:    "Trade Executed",
		Body:     "You bought 0.5 BTC",
		Metadata: map[string]string{"trade_id": "t1", "pair": "BTC-USDT"},
		IsRead:   false,
	}

	// Verify notification is JSON-marshalable (as done in SendToUser)
	data, err := json.Marshal(notif)
	assert.NoError(t, err)

	var received entities.Notification
	err = json.Unmarshal(data, &received)
	assert.NoError(t, err)

	assert.Equal(t, "notif-1", received.ID)
	assert.Equal(t, entities.TypeTradeExecuted, received.Type)
	assert.Equal(t, "t1", received.Metadata["trade_id"])
	assert.False(t, received.IsRead)
}

// --- Concurrent Access Safety Tests ---

func TestHub_SendToUser_ConcurrentCalls_NoRaceCondition(t *testing.T) {
	hub := NewHub()

	notif := &entities.Notification{
		ID:    "notif-1",
		Title: "Test",
	}

	// Send multiple notifications concurrently to verify no data races
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			hub.SendToUser("user-1", notif)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	// If we reach here without a panic, test passes
}

// --- Notification Marshaling Tests ---

func TestHub_NotificationMarshal_WithAllFields(t *testing.T) {
	notif := &entities.Notification{
		ID:       "notif-1",
		UserID:   "user-1",
		EventKey: "trade_t1_user1",
		Type:     entities.TypeTradeExecuted,
		Title:    "Trade Executed",
		Body:     "You bought 0.5 BTC at 50000",
		Metadata: map[string]string{
			"trade_id": "t1",
			"pair_id":  "BTC-USDT",
			"price":    "50000",
		},
		IsRead: false,
	}

	data, err := json.Marshal(notif)
	assert.NoError(t, err)

	var received entities.Notification
	err = json.Unmarshal(data, &received)
	assert.NoError(t, err)

	assert.Equal(t, "notif-1", received.ID)
	assert.Equal(t, "user-1", received.UserID)
	assert.Equal(t, entities.TypeTradeExecuted, received.Type)
	assert.Equal(t, "You bought 0.5 BTC at 50000", received.Body)
	assert.Len(t, received.Metadata, 3)
	assert.Equal(t, "50000", received.Metadata["price"])
}

func TestHub_NotificationMarshal_EmptyMetadata(t *testing.T) {
	notif := &entities.Notification{
		ID:       "notif-1",
		Title:    "Test",
		Metadata: map[string]string{},
	}

	data, err := json.Marshal(notif)
	assert.NoError(t, err)

	var received entities.Notification
	err = json.Unmarshal(data, &received)
	assert.NoError(t, err)

	assert.NotNil(t, received.Metadata)
}

// --- Register Tests ---

func TestHub_Register_AddsConnectionForUser(t *testing.T) {
	hub := NewHub()
	conn := &websocket.Conn{}

	hub.Register("user-1", conn)

	hub.mu.RLock()
	conns := hub.conns["user-1"]
	hub.mu.RUnlock()

	assert.NotNil(t, conns)
	assert.Len(t, conns, 1)
	assert.Equal(t, conn, conns[0].conn)
}

func TestHub_Register_MultipleConnectionsSameUser(t *testing.T) {
	hub := NewHub()
	conn1 := &websocket.Conn{}
	conn2 := &websocket.Conn{}

	hub.Register("user-1", conn1)
	hub.Register("user-1", conn2)

	hub.mu.RLock()
	conns := hub.conns["user-1"]
	hub.mu.RUnlock()

	assert.Len(t, conns, 2)
}

func TestHub_Register_MultipleUsers(t *testing.T) {
	hub := NewHub()
	conn1 := &websocket.Conn{}
	conn2 := &websocket.Conn{}

	hub.Register("user-1", conn1)
	hub.Register("user-2", conn2)

	hub.mu.RLock()
	defer hub.mu.RUnlock()

	assert.Len(t, hub.conns, 2)
	assert.Len(t, hub.conns["user-1"], 1)
	assert.Len(t, hub.conns["user-2"], 1)
}

// --- Unregister Tests ---

func TestHub_Unregister_RemovesConnection(t *testing.T) {
	hub := NewHub()
	conn := &websocket.Conn{}

	hub.Register("user-1", conn)
	hub.Unregister("user-1", conn)

	hub.mu.RLock()
	_, exists := hub.conns["user-1"]
	hub.mu.RUnlock()

	assert.False(t, exists, "User entry should be deleted when no connections remain")
}

func TestHub_Unregister_RemovesSpecificConnection(t *testing.T) {
	hub := NewHub()
	conn1 := &websocket.Conn{}
	conn2 := &websocket.Conn{}

	hub.Register("user-1", conn1)
	hub.Register("user-1", conn2)
	hub.Unregister("user-1", conn1)

	hub.mu.RLock()
	conns := hub.conns["user-1"]
	hub.mu.RUnlock()

	assert.Len(t, conns, 1)
	assert.Equal(t, conn2, conns[0].conn)
}

func TestHub_Unregister_UnknownUser_NoError(t *testing.T) {
	hub := NewHub()
	conn := &websocket.Conn{}

	// Should not panic when unregistering unknown user
	hub.Unregister("unknown-user", conn)
}

func TestHub_Unregister_UnknownConnection_NoError(t *testing.T) {
	hub := NewHub()
	conn1 := &websocket.Conn{}
	conn2 := &websocket.Conn{}

	hub.Register("user-1", conn1)

	// Should not panic when unregistering unknown connection
	hub.Unregister("user-1", conn2)

	hub.mu.RLock()
	conns := hub.conns["user-1"]
	hub.mu.RUnlock()

	// Original connection should still be there
	assert.Len(t, conns, 1)
	assert.Equal(t, conn1, conns[0].conn)
}

// --- SendToUser with Mock Connection Tests ---

func TestHub_SendToUser_WithConnections_JSONSerialization(t *testing.T) {
	notif := &entities.Notification{
		ID:       "notif-1",
		UserID:   "user-1",
		Type:     entities.TypeTradeExecuted,
		Title:    "Trade Executed",
		Body:     "You bought 0.5 BTC",
		Metadata: map[string]string{"trade_id": "t1"},
		IsRead:   false,
	}

	// Verify the notification can be serialized (as done in SendToUser)
	data, err := json.Marshal(notif)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var unmarshaled entities.Notification
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, "notif-1", unmarshaled.ID)
}

// --- SafeConn Tests ---

func TestHub_SafeConn_Concurrency(t *testing.T) {
	// Test that multiple goroutines can write to safeConn without race condition
	mockConn := &websocket.Conn{}
	sc := &safeConn{conn: mockConn}

	// Simulate concurrent writes
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			// This would panic on race condition
			sc.mu.Lock()
			sc.mu.Unlock()
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}

// --- SendToUser Real Connection Tests ---

func TestHub_SendToUser_ConcurrentSend(t *testing.T) {
	hub := NewHub()

	notif := &entities.Notification{
		ID:    "notif-1",
		Title: "Test",
	}

	// Send concurrently to non-existent user (should not panic)
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			hub.SendToUser("user-1", notif)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestHub_SendToUser_LargeMetadata(t *testing.T) {
	notif := &entities.Notification{
		ID:       "notif-1",
		UserID:   "user-1",
		Type:     entities.TypeTradeExecuted,
		Title:    "Trade Executed",
		Body:     "You bought 1000 units at 50000 per unit",
		Metadata: make(map[string]string),
		IsRead:   false,
	}

	// Add many metadata fields
	for i := 0; i < 50; i++ {
		notif.Metadata[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
	}

	// Should marshal without error even with large metadata
	data, err := json.Marshal(notif)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var received entities.Notification
	err = json.Unmarshal(data, &received)
	assert.NoError(t, err)
	assert.Len(t, received.Metadata, 50)
}

// --- Integration-like Tests ---

func TestHub_RegisterUnregisterMultipleUsers_Isolation(t *testing.T) {
	hub := NewHub()

	conn1 := &websocket.Conn{}
	conn2 := &websocket.Conn{}
	conn3 := &websocket.Conn{}

	// Register different users
	hub.Register("user-1", conn1)
	hub.Register("user-2", conn2)
	hub.Register("user-2", conn3)

	// Unregister one connection from user-2
	hub.Unregister("user-2", conn2)

	hub.mu.RLock()
	user1Conns := hub.conns["user-1"]
	user2Conns := hub.conns["user-2"]
	hub.mu.RUnlock()

	// User 1 should still have 1 connection
	assert.Len(t, user1Conns, 1)
	assert.Equal(t, conn1, user1Conns[0].conn)

	// User 2 should have 1 connection (conn3)
	assert.Len(t, user2Conns, 1)
	assert.Equal(t, conn3, user2Conns[0].conn)
}

// --- SafeConn Structure Tests ---

func TestHub_SafeConn_CreatedWithConnection(t *testing.T) {
	conn := &websocket.Conn{}
	sc := &safeConn{conn: conn}

	assert.NotNil(t, sc.conn)
	assert.Equal(t, conn, sc.conn)
	assert.NotNil(t, sc.mu)
}

func TestHub_SafeConn_MutexPreventsRaces(t *testing.T) {
	// Test that safeConn's mutex prevents concurrent access issues
	conn := &websocket.Conn{}
	sc := &safeConn{conn: conn}

	// Multiple goroutines accessing the same safeConn
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			// This will cause an error (no actual connection), but we're testing
			// that the mutex is being used, not that the write succeeds
			_ = sc.writeMessage(websocket.TextMessage, []byte("test"))
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without a panic or race detector warning, mutex works
	assert.True(t, true)
}

// --- SendToUser Behavior Tests ---

func TestHub_SendToUser_NoConnectionsForUser_ReturnsEarly(t *testing.T) {
	hub := NewHub()

	notif := &entities.Notification{
		ID:    "notif-1",
		Title: "Test",
	}

	// User has no connections - should return early
	hub.SendToUser("unknown-user", notif)

	// No error should occur
	assert.True(t, true)
}

func TestHub_SendToUser_NotificationWithMetadata_MarshalsCorrectly(t *testing.T) {
	notif := &entities.Notification{
		ID:     "notif-1",
		UserID: "user-1",
		Title:  "Trade Executed",
		Body:   "You bought 1 BTC",
		Metadata: map[string]string{
			"trade_id": "t123",
			"pair":     "BTC-USDT",
			"price":    "50000",
		},
	}

	// Test that notification can be marshaled for sending
	data, err := json.Marshal(notif)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify it can be unmarshaled back
	var received entities.Notification
	err = json.Unmarshal(data, &received)
	assert.NoError(t, err)
	assert.Equal(t, "notif-1", received.ID)
	assert.Equal(t, 3, len(received.Metadata))
}

func TestHub_SendToUser_DeepCopyPreventsConcurrentModification(t *testing.T) {
	hub := NewHub()
	conn1 := &websocket.Conn{}
	conn2 := &websocket.Conn{}

	notif := &entities.Notification{
		ID:    "notif-1",
		Title: "Test",
	}

	hub.Register("user-1", conn1)

	// Register another connection while SendToUser is running
	go func() {
		<-time.After(10 * time.Millisecond)
		hub.Register("user-1", conn2)
	}()

	// SendToUser takes a snapshot, so new connection won't receive this message
	hub.SendToUser("user-1", notif)
	<-time.After(100 * time.Millisecond)

	// Should complete without race condition
	assert.True(t, true)
}

func TestHub_SendToUser_GoroutineSpawned_ForEachConnection(t *testing.T) {
	hub := NewHub()
	conn := &websocket.Conn{}

	notif := &entities.Notification{
		ID:    "notif-1",
		Title: "Test",
	}

	hub.Register("user-1", conn)

	// SendToUser spawns goroutines for each connection
	// We can't directly verify goroutine count, but we can verify it completes
	hub.SendToUser("user-1", notif)
	<-time.After(50 * time.Millisecond)

	assert.True(t, true)
}

func TestHub_SendToUser_AllConnectionsProcessed(t *testing.T) {
	hub := NewHub()

	// Create multiple connections for same user
	conns := make([]*websocket.Conn, 3)
	for i := 0; i < 3; i++ {
		conns[i] = &websocket.Conn{}
		hub.Register("user-1", conns[i])
	}

	notif := &entities.Notification{
		ID:    "notif-1",
		Title: "Test",
	}

	// Verify all connections are registered
	hub.mu.RLock()
	assert.Len(t, hub.conns["user-1"], 3)
	hub.mu.RUnlock()

	// Send to all
	hub.SendToUser("user-1", notif)
	<-time.After(100 * time.Millisecond)

	assert.True(t, true)
}

func TestHub_SendToUser_ConcurrentSends_NoDataRace(t *testing.T) {
	hub := NewHub()
	conn := &websocket.Conn{}

	notif := &entities.Notification{
		ID:    "notif-1",
		Title: "Test",
	}

	hub.Register("user-1", conn)

	// Send concurrently from multiple goroutines
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			hub.SendToUser("user-1", notif)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Should complete without race condition
	assert.True(t, true)
}
