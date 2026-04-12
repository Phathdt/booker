package ws

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gofiber/contrib/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	sendBufferSize = 256
)

// Client represents a single WebSocket connection.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func NewClient(conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, sendBufferSize),
	}
}

// ReadPump reads messages from the WebSocket connection.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var sub SubscribeMsg
		if err := json.Unmarshal(msg, &sub); err != nil {
			c.sendError("invalid message format")
			continue
		}

		if sub.Channel != "ticker" && sub.Channel != "trades" {
			c.sendError("unknown channel: " + sub.Channel)
			continue
		}

		switch sub.Op {
		case "subscribe":
			c.hub.Subscribe(c, sub.Channel, sub.Pair)
		case "unsubscribe":
			c.hub.Unsubscribe(c, sub.Channel, sub.Pair)
		default:
			c.sendError("unknown op: " + sub.Op)
		}
	}
}

// WritePump writes messages to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) sendError(msg string) {
	errMsg := WSMessage{Type: "error", Msg: msg}
	data, err := json.Marshal(errMsg)
	if err != nil {
		slog.Error("failed to marshal error message", "error", err.Error())
		return
	}
	select {
	case c.send <- data:
	default:
	}
}
