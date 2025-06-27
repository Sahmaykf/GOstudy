package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
	}
	return server

}
func (this *Server) Handler(conn net.Conn) {
	fmt.Println("Linked Succeeded")
}
func (this *Server) Start() {
	fmt.Println("Start() 已运行") // 👈 先确保启动函数被调用
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.listen err:", err)
		return
	}
	fmt.Printf("服务器已启动，监听中：%s:%d\n", this.Ip, this.Port)
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		go this.Handler(conn)
	}
}
