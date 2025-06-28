package main

import "net"

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	go user.ListenMessage()
	return user
}

func (now *User) Online() {

	now.server.maplock.Lock()
	now.server.OnlineMap[now.Name] = now
	now.server.maplock.Unlock()
	now.server.BroadCast(now, "上线")
}

func (now *User) Offline() {
	now.server.maplock.Lock()
	delete(now.server.OnlineMap, now.Name)
	now.server.maplock.Unlock()
	now.server.BroadCast(now, "下线")
}

func (now *User) DoMessage(msg string) {
	now.server.BroadCast(now, msg)
}

func (now *User) ListenMessage() {
	for {
		msg := <-now.C
		now.conn.Write([]byte(msg + "\n"))
	}
}
