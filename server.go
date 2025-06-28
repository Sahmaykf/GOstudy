package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	OnlineMap map[string]*User
	maplock   sync.RWMutex

	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server

}
func (now *Server) ListenMessager() {
	for {
		msg := <-now.Message
		now.maplock.Lock()
		for _, cil := range now.OnlineMap {
			cil.C <- msg
		}
		now.maplock.Unlock()
	}
}
func (now *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	now.Message <- sendMsg
}
func (now *Server) Handler(conn net.Conn) {
	//fmt.Println("Linked Succeeded")

	user := NewUser(conn, now)
	//广播信息
	user.Online()
	//用户发信息 读进来
	go func() {
		buf := make([]byte, 4096)
		for {

			conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			n, err := conn.Read(buf)
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					user.sendMsg("你被踢了")
					user.Offline()
					return
				}
				if err == io.EOF {
					user.Offline()
					return
				}
				fmt.Println("Conn Read err:", err)
				return
			}
			if n > 0 {
				msg := string(buf[:n-1])
				//now.BroadCast(user, msg)
				//用户处理信息
				user.DoMessage(msg)
			}

		}
	}()
	select {}
}
func (now *Server) Start() {
	fmt.Println("Start() 已运行") // 先确保启动函数被调用
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", now.Ip, now.Port))
	if err != nil {
		fmt.Println("net.listen err:", err)
		return
	}
	fmt.Printf("服务器已启动，监听中：%s:%d\n", now.Ip, now.Port)
	defer listener.Close()
	go now.ListenMessager()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		go now.Handler(conn)
	}
}
