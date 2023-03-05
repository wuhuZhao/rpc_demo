# go 简易rpc框架编写
RPC(Remote Procedure Call Protocol)远程过程调用协议。 一个通俗的描述是：客户端在不知道调用细节的情况下，调用存在于远程计算机上的某个对象，就像调用本地应用程序中的对象一样。 比较正式的描述是：一种通过网络从远程计算机程序上请求服务，而不需要了解底层网络技术的协议
从使用的方面来说，服务端和客户端通过TCP/UDP/HTTP等通讯协议通讯，在通讯的时候客户端指定好服务端的方法、参数等信息通过序列化传送到服务端，服务端可以通过已有的元信息找到需要调用的方法，然后完成一次调用后序列化返回给客户端(rpc更多的是指服务与服务之间的通信，可以使用效率更高的协议和序列化格式去进行，并且可以进行有效的负载均衡和熔断超时等，因此跟前后端之间的web的交互概念上是有点不一样的)
用一张简单的图来表示

![image.png](https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/fd68bc0c6bbc40d49bb9c75ae6c90d52~tplv-k3u1fbpfcp-watermark.image?)
## 开始
本文只实现一个rpc框架基本的功能，不对性能做保证，因此尽量使用go原生自带的net/json库等进行操作，对使用方面不做stub（偷懒，只使用简单的json格式指定需要调用的方法），用最简单的方式实现一个简易rpc框架，也不保证超时调用和服务发现等集成的逻辑，[服务发现可以参考下文](https://juejin.cn/post/7172818468415209508)
本文代码地址(https://github.com/wuhuZhao/rpc_demo)
### 实现两点之间的通讯(transport)
本段先实现两端之间的通讯，只确保两个端之间能互相通讯即可 
`server.go`

```go
package server

import (
	"fmt"
	"log"
	"net"
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
			go func() {
				buf := make([]byte, 1024)
				for {
					idx, err := conn.Read(buf)
					if err != nil {
						panic(err)
					}
					if len(buf) == 0 {
						continue
					}
					// todo 等序列化的信息
					log.Printf("[conn: %v] get data: %v\n", conn.RemoteAddr(), string(buf[:idx]))

				}
			}()
		}
	}()

}

// Close: 关闭服务监听
func (s *Server) Close() error {
	return s.ls.Close()
}


// Close: 关闭服务监听
func (s *Server) Close() error {
	return s.ls.Close()
}
```

`client.go`

```go
package client

import (
	"fmt"
	"log"
	"net"
	"unsafe"
)

type Client struct {
	port int
	conn net.Conn
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

}

func (c *Client) Send(statement string) error {
	_, err := c.conn.Write(*(*[]byte)(unsafe.Pointer(&statement)))
	if err != nil {
		panic(err)
	}
	return nil
}

// Close: use to close connection
func (c *Client) Close() error {
	return c.conn.Close()
}


```
使用main.go做测试
`main.go`

```go
package main

import (
	"rpc_demo/internal/client"
	"rpc_demo/internal/server"
	"time"
)

func main() {
	s := server.NewServer(9999)
	s.Start()
	time.Sleep(5 * time.Second)
	c := client.NewClient(9999)
	c.Send("this is a test\n")
	time.Sleep(5 * time.Second)
}
```

执行一次`main.go`, `go run main.go`
```bash
2023/03/05 14:39:11 server [127.0.0.1:9999] start....
2023/03/05 14:39:16 [conn: 127.0.0.1:59126] get data: this is a test
```

可以证明第一部分的任务已经完成，可以实现两端之间的通讯了

### 实现反射调用已注册的方法
实现了双端的通信以后，我们在`internal.go`里实现两个方法，一个是注册，一个是调用，因为go有运行时的反射，所以我们使用反射去注册每一个需要调用到的方法，然后提供全局唯一的函数名，让client端可以实现指定方法的调用

`internal.go`

```go
package internal

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// 全局唯一
var GlobalMethod = &Method{methods: map[string]reflect.Value{}}

type Method struct {
	methods map[string]reflect.Value
}

func (m *Method) register(impl interface{}) error {
	pl := reflect.ValueOf(impl)
	if pl.Kind() != reflect.Func {
		return errors.New("impl should be function")
	}
	// 获取函数名
	methodName := runtime.FuncForPC(pl.Pointer()).Name()
	if len(strings.Split(methodName, ".")) < 1 {
		return errors.New("invalid function name")
	}
	lastFuncName := strings.Split(methodName, ".")[1]
	m.methods[lastFuncName] = pl
	fmt.Printf("methods: %v\n", m.methods)
	return nil
}

func (m *Method) call(methodName string, callParams ...interface{}) ([]interface{}, error) {
	fn, ok := m.methods[methodName]
	if !ok {
		return nil, errors.New("impl method not found! Please Register first")
	}
	in := make([]reflect.Value, len(callParams))
	for i := 0; i < len(callParams); i++ {
		in[i] = reflect.ValueOf(callParams[i])
	}
	res := fn.Call(in)
	out := make([]interface{}, len(res))
	for i := 0; i < len(res); i++ {
		out[i] = res[i].Interface()
	}
	return out, nil
}

func Call(methodName string, callParams ...interface{}) ([]interface{}, error) {
	return GlobalMethod.call(methodName, callParams...)
}

func Register(impl interface{}) error {
	return GlobalMethod.register(impl)
}

```
在单测里测试一下这个注册和调用的功能`internal_test.go`

```go
package internal

import (
	"testing"
)

func Sum(a, b int) int {
	return a + b
}
func TestRegister(t *testing.T) {
	err := Register(Sum)
	if err != nil {
		t.Fatalf("err: %v\n", err)
	}
	t.Logf("test success\n")
}

func TestCall(t *testing.T) {
	TestRegister(t)
	result, err := Call("Sum", 1, 2)
	if err != nil {
		t.Fatalf("err: %v\n", err)
	}
	if len(result) != 1 {
		t.Fatalf("len(result) is not equal to 1\n")
	}
	t.Logf("Sum(1,2) = %d\n", result[0].(int))
	if err := recover(); err != nil {
		t.Fatalf("%v\n", err)
	}
}
```
执行调用
```bash
/usr/local/go/bin/go test -timeout 30s -run ^TestCall$ rpc_demo/internal -v
```

```bash
Running tool: /usr/local/go/bin/go test -timeout 30s -run ^TestCall$ rpc_demo/internal -v

=== RUN   TestCall
methods: map[Sum:<func(int, int) int Value>]
    /root/go/src/juejin_demo/rpc_demo/internal/internal_test.go:15: test success
    /root/go/src/juejin_demo/rpc_demo/internal/internal_test.go:27: Sum(1,2) = 3
--- PASS: TestCall (0.00s)
PASS
ok  	rpc_demo/internal	0.002s
```

可以看到这个注册和调用的过程已经实现并且达到指定方法调用的作用

### 设计struct完整表达一次完整的rpc调用，并且封装json库中的Decoder和Encoder，完成序列化和反序列化

`internal.go`
```go
type RpcRequest struct {
	MethodName string
	Params     []interface{}
}

type RpcResponses struct {
	Returns []interface{}
	Err     error
}

```

`transport.go`考虑可以对接更多的格式，所以抽象了一层进行使用（demo肯定没有更多格式了）
```go
package transport

// Transport: 序列化格式的抽象层，从connection中读取数据序列化并且反序列化到connection中
type Transport interface {
	Decode(v interface{}) error
	Encode(v interface{}) error
	Close()
}

```

`json_transport.go`
```go
package transport

import (
	"encoding/json"
	"net"
)

var _ Transport = (*JSONTransport)(nil)

type JSONTransport struct {
	encoder *json.Encoder
	decoder *json.Decoder
}

// NewJSONTransport: 负责读取和写入conn
func NewJSONTransport(conn net.Conn) *JSONTransport {
	return &JSONTransport{json.NewEncoder(conn), json.NewDecoder(conn)}
}

// Decode: use json package to decode
func (t *JSONTransport) Decode(v interface{}) error {
	if err := t.decoder.Decode(v); err != nil {
		return err
	}
	return nil
}

// Encode: use json package to encode
func (t *JSONTransport) Encode(v interface{}) error {
	if err := t.encoder.Encode(v); err != nil {
		return err
	}
	return nil
}

// Close: not implement
func (dec *JSONTransport) Close() {

}

```
然后我们将服务端和客户端的逻辑进行修改，改成通过上面两个结构体进行通信，然后返回一次调用
`server.go`
```go
//...
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
        //...
```

`client.go`
```go
// ...
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
// ...
```

`main.go`
```go
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

```

接下来我们运行一下main
```bash
[root@hecs-74066 rpc_demo]# go run main.go 
2023/03/05 14:39:11 server [127.0.0.1:9999] start....
2023/03/05 14:39:16 [conn: 127.0.0.1:59126] get data: this is a test

[root@hecs-74066 rpc_demo]# go run main.go 
2023/03/05 21:53:41 server [127.0.0.1:9999] start....
2023/03/05 21:53:46 [client] create request to invoke server: {Join [aaaaa bbbbb ccccccccc end]}
2023/03/05 21:53:46 [server] get request: &{Join [aaaaa bbbbb ccccccccc end]}
2023/03/05 21:53:46 [server] invoke method: [aaaaabbbbbcccccccccend]
2023/03/05 21:53:46 [client] get response from server: &{[aaaaabbbbbcccccccccend] <nil>}
2023/03/05 21:53:46 [main] get a response from server: aaaaabbbbbcccccccccend
```

## 总结(自我pua)
这样我们就实现了一个简单的rpc框架了，符合最简单的架构图，从client->序列化请求->transport -> 反序列化 ->server然后从server->序列化请求->transport->反序列化请求->client。当然从可用性的角度来说是差远了，没有实现stub代码，也没有idl的实现，导致所有的注册方法都是硬编码，可用性不高，而且没有[集成服务发现(可以参考我的另一篇文章去集成)](https://juejin.cn/post/7172818468415209508)和熔断等功能，也没用[中间件(也是我的另一篇文章)](https://juejin.cn/post/7169215426423947278)和超时等丰富的功能在里面，并且最近看了不少rpc框架的源码，感觉这个demo的设计也差远了。不过因为时间问题和代码的复杂性问题（单纯懒），起码算是实现了一个简单的rpc框架。

推荐一些比较好的框架实现
1. [kitex](https://www.cloudwego.io/zh/docs/kitex/overview/)
2. dubbo
3. grpc
4. thrift