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

// TestHub_Register_Success tests successful client registration
func TestHub_Register_Success(t *testing.T) {
	hub := NewHub()
	go hub.Run(context.Background())

	client := &Client{hub: hub, conn: nil, send: make(chan []byte, sendBufferSize)}

	hub.Register(client)
	time.Sleep(100 * time.Millisecond) // Let event loop process

	hub.mu.RLock()
	assert.True(t, hub.clients[client])
	hub.mu.RUnlock()
}

// TestHub_Unregister_Success tests successful client unregistration
func TestHub_Unregister_Success(t *testing.T) {
	hub := NewHub()
	ctx := context.Background()
	go hub.Run(ctx)

	client := &Client{hub: hub, conn: nil, send: make(chan []byte, sendBufferSize)}

	hub.Register(client)
	time.Sleep(100 * time.Millisecond)

	hub.Unregister(client)
	time.Sleep(100 * time.Millisecond)

	hub.mu.RLock()
	assert.False(t, hub.clients[client])
	hub.mu.RUnlock()
}

// TestHub_Subscribe_Success tests successful subscription to a channel:pair
func TestHub_Subscribe_Success(t *testing.T) {
	hub := NewHub()
	client := &Client{hub: hub, conn: nil, send: make(chan []byte, sendBufferSize)}

	hub.Subscribe(client, "ticker", "BTC_USDT")

	hub.mu.RLock()
	key := subKey{channel: "ticker", pair: "BTC_USDT"}
	assert.True(t, hub.subs[key][client])
	hub.mu.RUnlock()
}

// TestHub_Subscribe_MultipleChannels tests subscribing to multiple channels
func TestHub_Subscribe_MultipleChannels(t *testing.T) {
	hub := NewHub()
	client := &Client{hub: hub, conn: nil, send: make(chan []byte, sendBufferSize)}

	hub.Subscribe(client, "ticker", "BTC_USDT")
	hub.Subscribe(client, "trades", "BTC_USDT")
	hub.Subscribe(client, "ticker", "ETH_USDT")

	hub.mu.RLock()
	defer hub.mu.RUnlock()

	assert.True(t, hub.subs[subKey{channel: "ticker", pair: "BTC_USDT"}][client])
	assert.True(t, hub.subs[subKey{channel: "trades", pair: "BTC_USDT"}][client])
	assert.True(t, hub.subs[subKey{channel: "ticker", pair: "ETH_USDT"}][client])
}

// TestHub_Unsubscribe_Success tests successful unsubscription
func TestHub_Unsubscribe_Success(t *testing.T) {
	hub := NewHub()
	client := &Client{hub: hub, conn: nil, send: make(chan []byte, sendBufferSize)}

	hub.Subscribe(client, "ticker", "BTC_USDT")
	hub.Unsubscribe(client, "ticker", "BTC_USDT")

	hub.mu.RLock()
	defer hub.mu.RUnlock()

	key := subKey{channel: "ticker", pair: "BTC_USDT"}
	_, exists := hub.subs[key]
	assert.False(t, exists)
}

// TestHub_Unsubscribe_RemovesEmptySubscriptionMap tests cleanup of empty subscription maps
func TestHub_Unsubscribe_RemovesEmptySubscriptionMap(t *testing.T) {
	hub := NewHub()
	client := &Client{hub: hub, conn: nil, send: make(chan []byte, sendBufferSize)}

	hub.Subscribe(client, "ticker", "BTC_USDT")
	assert.True(t, len(hub.subs) > 0)

	hub.Unsubscribe(client, "ticker", "BTC_USDT")

	hub.mu.RLock()
	defer hub.mu.RUnlock()

	key := subKey{channel: "ticker", pair: "BTC_USDT"}
	_, exists := hub.subs[key]
	assert.False(t, exists)
}

// TestHub_Unsubscribe_NonexistentSubscription tests unsubscribing from non-existent subscription
func TestHub_Unsubscribe_NonexistentSubscription(t *testing.T) {
	hub := NewHub()
	client := &Client{hub: hub, conn: nil, send: make(chan []byte, sendBufferSize)}

	// Should not panic
	hub.Unsubscribe(client, "ticker", "BTC_USDT")

	hub.mu.RLock()
	defer hub.mu.RUnlock()

	assert.Equal(t, 0, len(hub.subs))
}

// TestHub_BroadcastTicker_Success tests broadcasting ticker updates
func TestHub_BroadcastTicker_Success(t *testing.T) {
	hub := NewHub()
	client := &Client{hub: hub, conn: nil, send: make(chan []byte, sendBufferSize)}

	hub.mu.Lock()
	hub.clients[client] = true
	hub.mu.Unlock()

	hub.Subscribe(client, "ticker", "BTC_USDT")

	tickerData := TickerData{
		Open:      "45000.00",
		High:      "46000.00",
		Low:       "44000.00",
		Close:     "45500.00",
		Volume:    "1000.00",
		ChangePct: "1.11",
		LastPrice: "45500.00",
		Timestamp: time.Now().Unix(),
	}

	hub.BroadcastTicker("BTC_USDT", tickerData)
	time.Sleep(100 * time.Millisecond)

	var msg WSMessage
	select {
	case data := <-client.send:
		err := json.Unmarshal(data, &msg)
		require.NoError(t, err)
		assert.Equal(t, "ticker", msg.Type)
		assert.Equal(t, "BTC_USDT", msg.Pair)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

// TestHub_BroadcastTrade_Success tests broadcasting trade events
func TestHub_BroadcastTrade_Success(t *testing.T) {
	hub := NewHub()
	client := &Client{hub: hub, conn: nil, send: make(chan []byte, sendBufferSize)}

	hub.mu.Lock()
	hub.clients[client] = true
	hub.mu.Unlock()

	hub.Subscribe(client, "trades", "BTC_USDT")

	tradeData := TradeData{
		TradeID:   "trade-123",
		Price:     "45500.00",
		Quantity:  "0.5",
		Timestamp: time.Now().Unix(),
	}

	hub.BroadcastTrade("BTC_USDT", tradeData)
	time.Sleep(100 * time.Millisecond)

	var msg WSMessage
	select {
	case data := <-client.send:
		err := json.Unmarshal(data, &msg)
		require.NoError(t, err)
		assert.Equal(t, "trade", msg.Type)
		assert.Equal(t, "BTC_USDT", msg.Pair)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

// TestHub_BroadcastToMultipleClients tests broadcasting to multiple subscribers
func TestHub_BroadcastToMultipleClients(t *testing.T) {
	hub := NewHub()
	client1 := &Client{hub: hub, conn: nil, send: make(chan []byte, sendBufferSize)}
	client2 := &Client{hub: hub, conn: nil, send: make(chan []byte, sendBufferSize)}

	hub.mu.Lock()
	hub.clients[client1] = true
	hub.clients[client2] = true
	hub.mu.Unlock()

	hub.Subscribe(client1, "ticker", "BTC_USDT")
	hub.Subscribe(client2, "ticker", "BTC_USDT")

	tickerData := TickerData{
		Open:      "45000.00",
		High:      "46000.00",
		Low:       "44000.00",
		Close:     "45500.00",
		Volume:    "1000.00",
		ChangePct: "1.11",
		LastPrice: "45500.00",
		Timestamp: time.Now().Unix(),
	}

	hub.BroadcastTicker("BTC_USDT", tickerData)
	time.Sleep(100 * time.Millisecond)

	// Both clients should receive the message
	select {
	case data := <-client1.send:
		var msg WSMessage
		err := json.Unmarshal(data, &msg)
		require.NoError(t, err)
		assert.Equal(t, "ticker", msg.Type)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout for client1")
	}

	select {
	case data := <-client2.send:
		var msg WSMessage
		err := json.Unmarshal(data, &msg)
		require.NoError(t, err)
		assert.Equal(t, "ticker", msg.Type)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout for client2")
	}
}

// TestHub_BroadcastOnlyToSubscribers tests that broadcasts only reach subscribers
func TestHub_BroadcastOnlyToSubscribers(t *testing.T) {
	hub := NewHub()
	client1 := &Client{hub: hub, conn: nil, send: make(chan []byte, sendBufferSize)}
	client2 := &Client{hub: hub, conn: nil, send: make(chan []byte, sendBufferSize)}

	hub.mu.Lock()
	hub.clients[client1] = true
	hub.clients[client2] = true
	hub.mu.Unlock()

	// Only subscribe client1 to ticker:BTC_USDT
	hub.Subscribe(client1, "ticker", "BTC_USDT")
	// client2 subscribes to different pair
	hub.Subscribe(client2, "ticker", "ETH_USDT")

	tickerData := TickerData{
		Open:      "45000.00",
		High:      "46000.00",
		Low:       "44000.00",
		Close:     "45500.00",
		Volume:    "1000.00",
		ChangePct: "1.11",
		LastPrice: "45500.00",
		Timestamp: time.Now().Unix(),
	}

	hub.BroadcastTicker("BTC_USDT", tickerData)
	time.Sleep(100 * time.Millisecond)

	// client1 should receive
	select {
	case <-client1.send:
		// Expected
	case <-time.After(1 * time.Second):
		t.Fatal("client1 should receive message")
	}

	// client2 should NOT receive
	select {
	case <-client2.send:
		t.Fatal("client2 should not receive message for BTC_USDT")
	case <-time.After(100 * time.Millisecond):
		// Expected
	}
}

// TestHub_BroadcastToUnregisteredClient tests behavior when broadcast client not registered
func TestHub_BroadcastToUnregisteredClient(t *testing.T) {
	hub := NewHub()
	client := &Client{hub: hub, conn: nil, send: make(chan []byte, sendBufferSize)}

	// Subscribe but don't register the client
	hub.Subscribe(client, "ticker", "BTC_USDT")

	tickerData := TickerData{
		Open:      "45000.00",
		High:      "46000.00",
		Low:       "44000.00",
		Close:     "45500.00",
		Volume:    "1000.00",
		ChangePct: "1.11",
		LastPrice: "45500.00",
		Timestamp: time.Now().Unix(),
	}

	// Should not panic or cause issues
	hub.BroadcastTicker("BTC_USDT", tickerData)
	time.Sleep(100 * time.Millisecond)

	// Client should not receive (not registered)
	select {
	case <-client.send:
		t.Fatal("unregistered client should not receive")
	case <-time.After(100 * time.Millisecond):
		// Expected
	}
}

// TestHub_UnregisterRemovesFromAllSubscriptions tests that unregister removes client from all subs
func TestHub_UnregisterRemovesFromAllSubscriptions(t *testing.T) {
	hub := NewHub()
	ctx := context.Background()
	go hub.Run(ctx)

	client := &Client{hub: hub, conn: nil, send: make(chan []byte, sendBufferSize)}

	hub.Register(client)
	time.Sleep(100 * time.Millisecond)

	// Subscribe to multiple channels
	hub.Subscribe(client, "ticker", "BTC_USDT")
	hub.Subscribe(client, "trades", "BTC_USDT")
	hub.Subscribe(client, "ticker", "ETH_USDT")

	// Verify subscriptions
	hub.mu.RLock()
	initialSubCount := len(hub.subs)
	hub.mu.RUnlock()
	assert.Greater(t, initialSubCount, 0)

	// Unregister should remove from all subscriptions
	hub.Unregister(client)
	time.Sleep(100 * time.Millisecond)

	hub.mu.RLock()
	defer hub.mu.RUnlock()

	for _, subs := range hub.subs {
		assert.False(t, subs[client])
	}
}

// TestHub_ConcurrentSubscribeUnsubscribe tests concurrent operations
func TestHub_ConcurrentSubscribeUnsubscribe(t *testing.T) {
	hub := NewHub()
	client := &Client{hub: hub, conn: nil, send: make(chan []byte, sendBufferSize)}

	var wg sync.WaitGroup
	pairs := []string{"BTC_USDT", "ETH_USDT", "XRP_USDT", "ADA_USDT"}

	// Concurrent subscriptions
	for _, pair := range pairs {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			hub.Subscribe(client, "ticker", p)
		}(pair)
	}
	wg.Wait()

	hub.mu.RLock()
	assert.Equal(t, len(pairs), len(hub.subs))
	hub.mu.RUnlock()

	// Concurrent unsubscriptions
	for _, pair := range pairs {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			hub.Unsubscribe(client, "ticker", p)
		}(pair)
	}
	wg.Wait()

	hub.mu.RLock()
	assert.Equal(t, 0, len(hub.subs))
	hub.mu.RUnlock()
}

// TestHub_NewHub_InitializesCorrectly tests that NewHub creates correct empty state
func TestHub_NewHub_InitializesCorrectly(t *testing.T) {
	hub := NewHub()

	assert.NotNil(t, hub.clients)
	assert.NotNil(t, hub.subs)
	assert.NotNil(t, hub.register)
	assert.NotNil(t, hub.unregister)
	assert.Equal(t, 0, len(hub.clients))
	assert.Equal(t, 0, len(hub.subs))
}

// TestHub_Run_ContextCancellation tests that Run exits on context cancellation
func TestHub_Run_ContextCancellation(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		hub.Run(ctx)
		done <- struct{}{}
	}()

	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(1 * time.Second):
		t.Fatal("Run should exit on context cancellation")
	}
}
