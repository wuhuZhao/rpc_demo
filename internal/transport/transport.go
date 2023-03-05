package transport

// Transport: 序列化格式的抽象层，从connection中读取数据序列化并且反序列化到connection中
type Transport interface {
	Decode(v interface{}) error
	Encode(v interface{}) error
	Close()
}
