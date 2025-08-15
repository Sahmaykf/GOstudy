package main

import (
	"net"
	"time"

	"github.com/gorilla/websocket"
)

// Conn abstracts net.Conn and websocket connections
// to provide a common interface for server handlers.
type Conn interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Close() error
	RemoteAddr() net.Addr
	SetReadDeadline(t time.Time) error
}

// WSConn wraps a websocket connection to satisfy the Conn interface.
type WSConn struct {
	*websocket.Conn
}

func (c *WSConn) Read(p []byte) (int, error) {
	_, msg, err := c.ReadMessage()
	if err != nil {
		return 0, err
	}
	n := copy(p, msg)
	return n, nil
}

func (c *WSConn) Write(p []byte) (int, error) {
	if err := c.WriteMessage(websocket.TextMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}
