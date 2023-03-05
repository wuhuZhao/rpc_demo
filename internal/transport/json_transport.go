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
