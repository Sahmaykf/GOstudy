package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// StartWebSocket starts a websocket server on the given port.
func (s *Server) StartWebSocket(port int) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		go s.Handler(&WSConn{conn})
	})
	addr := fmt.Sprintf("%s:%d", s.Ip, port)
	fmt.Printf("WebSocket服务器已启动，监听中：%s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Println("WebSocket server err:", err)
	}
}
