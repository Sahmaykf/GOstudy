package main

import (
	"fmt"
	"net"
	"sync"
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
	//用户上线 加map
	user := NewUser(conn)
	now.maplock.Lock()
	now.OnlineMap[user.Name] = user
	now.maplock.Unlock()

	//广播信息
	now.BroadCast(user, "上线")
}
func (now *Server) Start() {
	fmt.Println("Start() 已运行") // 👈 先确保启动函数被调用
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
