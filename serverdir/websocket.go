package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (s *Server) StartWebSocket(port int) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Upgrader err: ", err)
			return
		}
		go s.Handler(&WSConn{conn})
	})
	s.ensureMessageListener()
	addr := fmt.Sprintf("%s:%d", s.Ip, port)
	fmt.Printf("WebSocket服务器已启动，监听中：%s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Println("WebSocket server err:", err)
	}
}
