package builtin

import (
	jsoniter "github.com/json-iterator/go"
)

type Codec interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type defaultCodec struct{}

func (c *defaultCodec) Marshal(v interface{}) ([]byte, error) {
	return jsoniter.Marshal(v)
}

func (c *defaultCodec) Unmarshal(data []byte, v interface{}) error {
	return jsoniter.Unmarshal(data, v)
}
