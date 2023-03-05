package client

import (
	"fmt"
	"log"
	"net"
	"rpc_demo/internal"
	"rpc_demo/internal/transport"
)

type Client struct {
	port int
	conn net.Conn
	tsp  transport.Transport
}

func NewClient(port int) *Client {
	c := &Client{port: port}
	c.init()
	return c
}

// init: initialize tcp client
func (c *Client) init() {
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", c.port))
	if err != nil {
		panic(err)
	}
	c.conn = conn
	c.tsp = transport.NewJSONTransport(conn)
}

// Call: remote invoke
func (c *Client) Call(methodName string, params ...interface{}) (res *internal.RpcResponses) {
	request := internal.RpcRequest{MethodName: methodName, Params: params}
	log.Printf("[client] create request to invoke server: %v\n", request)
	err := c.tsp.Encode(request)
	if err != nil {
		panic(err)
	}
	res = &internal.RpcResponses{}
	if err := c.tsp.Decode(res); err != nil {
		panic(err)
	}
	log.Printf("[client] get response from server: %v\n", res)
	return res
}

// Close: use to close connection
func (c *Client) Close() error {
	return c.conn.Close()
}
