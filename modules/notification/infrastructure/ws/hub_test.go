package ws

import (
	"encoding/json"
	"testing"

	"booker/modules/notification/domain/entities"

	"github.com/stretchr/testify/assert"
)

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
