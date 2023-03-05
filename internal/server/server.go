package server

import (
	"fmt"
	"log"
	"net"
	"rpc_demo/internal"
	"rpc_demo/internal/transport"
)

// Server: transport底层实现，通过Server去接受客户端的字节流
type Server struct {
	ls   net.Listener
	port int
}

// NewServer: 根据端口创建一个server
func NewServer(port int) *Server {
	s := &Server{port: port}
	s.init()
	return s
}

// init: 初始化服务端连接
func (s *Server) init() {
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", s.port))
	if err != nil {
		panic(err)
	}
	s.ls = l
}

// Start: 启动服务端的端口监听，采取一个conn一个g的模型，没有使用reactor等高性能模型
func (s *Server) Start() {
	go func() {
		log.Printf("server [%s] start....", s.ls.Addr().String())
		for {
			conn, err := s.ls.Accept()
			if err != nil {
				panic(err)
			}
			tsp := transport.NewJSONTransport(conn)
			go func() {
				for {
					request := &internal.RpcRequest{}
					err := tsp.Decode(request)
					if err != nil {
						panic(err)
					}
					log.Printf("[server] get request: %v\n", request)
					result, err := internal.Call(request.MethodName, request.Params...)
					log.Printf("[server] invoke method: %v\n", result)
					if err != nil {
						response := &internal.RpcResponses{Returns: nil, Err: err}
						tsp.Encode(response)
						continue
					}
					response := &internal.RpcResponses{Returns: result, Err: err}
					if err := tsp.Encode(response); err != nil {
						log.Printf("[server] encode response err: %v\n", err)
						continue
					}
				}
			}()
		}
	}()

}

// Close: 关闭服务监听
func (s *Server) Close() error {
	return s.ls.Close()
}
