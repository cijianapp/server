package model

import (
	"github.com/gorilla/websocket"
)

// WebConn type
type WebConn struct {
	WebSocket *websocket.Conn
}

// NewwebConn generate a new connect to one client
func NewwebConn(ws *websocket.Conn) *WebConn {
	wc := &WebConn{WebSocket: ws}

	return wc
}

// Pump means conn start work
func (c *WebConn) Pump() {
	ch := make(chan struct{})
	go func() {
		c.writePump()
		close(ch)
	}()
}

func (c *WebConn) writePump() {}
