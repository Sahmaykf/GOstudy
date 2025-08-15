package main

import "fmt"

func main() {
	server := NewServer("127.0.0.1", 8888)
	go server.StartWebSocket(8080)
	fmt.Println("main 已运行")
	server.Start()
}
