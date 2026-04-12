package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
)

// subscription key: "channel:pair" e.g. "ticker:BTC_USDT"
type subKey struct {
	channel string
	pair    string
}

// Hub manages WebSocket client subscriptions and broadcasts.
type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	subs       map[subKey]map[*Client]bool // channel:pair → clients
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		subs:       make(map[subKey]map[*Client]bool),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
	}
}

// Run starts the hub event loop. Exits when ctx is cancelled.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				// Remove from all subscriptions
				for key, clients := range h.subs {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.subs, key)
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

// Subscribe adds a client to a channel:pair subscription.
func (h *Hub) Subscribe(client *Client, channel, pair string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := subKey{channel: channel, pair: pair}
	if h.subs[key] == nil {
		h.subs[key] = make(map[*Client]bool)
	}
	h.subs[key][client] = true
}

// Unsubscribe removes a client from a channel:pair subscription.
func (h *Hub) Unsubscribe(client *Client, channel, pair string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := subKey{channel: channel, pair: pair}
	if clients, ok := h.subs[key]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.subs, key)
		}
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// BroadcastTicker sends a ticker update to all subscribers.
func (h *Hub) BroadcastTicker(pair string, data TickerData) {
	msg := WSMessage{Type: "ticker", Pair: pair, Data: data}
	h.broadcast("ticker", pair, msg)
}

// BroadcastTrade sends a trade event to all subscribers.
func (h *Hub) BroadcastTrade(pair string, data TradeData) {
	msg := WSMessage{Type: "trade", Pair: pair, Data: data}
	h.broadcast("trades", pair, msg)
}

func (h *Hub) broadcast(channel, pair string, msg WSMessage) {
	payload, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal WS message", "error", err.Error())
		return
	}

	h.mu.RLock()
	src := h.subs[subKey{channel: channel, pair: pair}]
	targets := make([]*Client, 0, len(src))
	for c := range src {
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	for _, client := range targets {
		select {
		case client.send <- payload:
		default:
			// Slow client — drop message
		}
	}
}
