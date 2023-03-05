package main

import (
	"log"
	"rpc_demo/internal"
	"rpc_demo/internal/client"
	"rpc_demo/internal/server"
	"strings"
	"time"
)

// Rpc方法的一个简易实现
func Join(a ...string) string {
	res := &strings.Builder{}
	for i := 0; i < len(a); i++ {
		res.WriteString(a[i])
	}
	return res.String()
}

func main() {
	internal.Register(Join)
	s := server.NewServer(9999)
	s.Start()
	time.Sleep(5 * time.Second)
	c := client.NewClient(9999)
	res := c.Call("Join", "aaaaa", "bbbbb", "ccccccccc", "end")
	if res.Err != nil {
		log.Printf("[main] get an error from server: %v\n", res.Err)
		return
	}
	log.Printf("[main] get a response from server: %v\n", res.Returns[0].(string))
	time.Sleep(5 * time.Second)
}
