package main

import (
	// "/Users/mima0000/Desktop/go/golang-IM-program/serverdir/data"
	"fmt"

	"github.com/Sahmaykf/GOstudy/serverdir/data"
)

func main() {
	db, err := data.InitDB()
	if err != nil {
		println("db err")
		return
	}
	_ = db
	server := NewServer("127.0.0.1", 8080)
	go server.StartWebSocket(8080)

	fmt.Println("main 已运行")
	//server.Start()
}
