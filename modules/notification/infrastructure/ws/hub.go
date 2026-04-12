package ws

import (
	"encoding/json"
	"log/slog"
	"sync"

	"booker/modules/notification/domain/entities"

	"github.com/gofiber/contrib/websocket"
)

// safeConn wraps a websocket.Conn with a write mutex for goroutine-safe writes.
type safeConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (sc *safeConn) writeMessage(msgType int, data []byte) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.conn.WriteMessage(msgType, data)
}

// Hub manages WebSocket connections per user.
type Hub struct {
	mu    sync.RWMutex
	conns map[string][]*safeConn
}

func NewHub() *Hub {
	return &Hub{
		conns: make(map[string][]*safeConn),
	}
}

// Register adds a WebSocket connection for a user.
func (h *Hub) Register(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[userID] = append(h.conns[userID], &safeConn{conn: conn})
}

// Unregister removes a WebSocket connection for a user.
func (h *Hub) Unregister(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conns := h.conns[userID]
	for i, sc := range conns {
		if sc.conn == conn {
			h.conns[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
	if len(h.conns[userID]) == 0 {
		delete(h.conns, userID)
	}
}

// SendToUser broadcasts a notification to all connections for a user.
func (h *Hub) SendToUser(userID string, notification *entities.Notification) {
	h.mu.RLock()
	src := h.conns[userID]
	if len(src) == 0 {
		h.mu.RUnlock()
		return
	}
	// Deep-copy slice to avoid race with Unregister
	conns := make([]*safeConn, len(src))
	copy(conns, src)
	h.mu.RUnlock()

	data, err := json.Marshal(notification)
	if err != nil {
		slog.Error("failed to marshal notification for WS", "error", err.Error())
		return
	}

	for _, sc := range conns {
		sc := sc // capture loop variable
		go func() {
			if err := sc.writeMessage(websocket.TextMessage, data); err != nil {
				slog.Warn("failed to send WS message", "user_id", userID, "error", err.Error())
			}
		}()
	}
}
