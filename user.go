package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
	done   chan struct{}
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		done:   make(chan struct{}),
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
	close(now.done)
	close(now.C)
	now.conn.Close()
}

func (now *User) sendMsg(msg string) {
	now.conn.Write([]byte(msg))
}

func (now *User) changeName(newName string) {
	now.server.maplock.Lock()
	delete(now.server.OnlineMap, now.Name)
	now.Name = newName
	now.server.OnlineMap[now.Name] = now
	now.server.maplock.Unlock()
	now.sendMsg("您的用户名更新为：" + now.Name + "\n")
}

func (now *User) DoMessage(msg string) {
	if msg == "who" {
		now.server.maplock.Lock()
		for _, user := range now.server.OnlineMap {
			now.sendMsg("[" + user.Addr + "]" + user.Name + ":" + "在线\n")
		}
		now.server.maplock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		_, ok := now.server.OnlineMap[newName]
		if ok {
			now.sendMsg("用户名重复了\n")
		} else {
			now.changeName(newName)
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			now.sendMsg("格式不正确,正确的格式为to|张三|message...\n")
			return
		}
		remoteUser, ok := now.server.OnlineMap[remoteName]

		if !ok {
			now.sendMsg("查无此人\n")
			return
		}
		content := strings.Split(msg, "|")[2]
		if content == "" {
			now.sendMsg("消息为空,请重试\n")
			return
		}
		remoteUser.sendMsg(now.Name + " say: " + content + "\n")
	} else {
		now.server.BroadCast(now, msg)
	}
}

func (now *User) ListenMessage() {
	for {
		select {
		case msg, ok := <-now.C:
			if !ok {
				return
			}
			now.conn.Write([]byte(msg + "\n"))
		case <-now.done:
			return
		}
	}
}
