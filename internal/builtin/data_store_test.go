package builtin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack"
)

func TestSnapshot(t *testing.T) {
	var ds = &defaultSharedDataStore{
		dataStore: dataStore{
			Data: map[string]interface{}{
				"a": 123,
				"b": "test",
			},
			NodeData: map[string]map[string]interface{}{
				"c": {
					"c1": 1,
					"c2": true,
				},
			},
		},
		Codec: &defaultCodec{},
	}
	_, e := ds.Marshal()
	assert.NoError(t, e)
}

type testCodec struct{}

func (t *testCodec) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (t *testCodec) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

func TestRestore(t *testing.T) {
	var codec = &testCodec{}
	var ds = &defaultSharedDataStore{
		dataStore: dataStore{
			Data: map[string]interface{}{
				"a": int64(123),
				"b": "test",
			},
			NodeData: map[string]map[string]interface{}{
				"c": {
					"c1": float64(1),
					"c2": true,
				},
			},
		},
		Codec: codec,
	}
	bs, _ := ds.Marshal()

	ds2 := NewSharedDataStore(WithCodec(codec))

	err := ds2.Unmarshal(bs)
	assert.NoError(t, err)
	for k, v := range ds.Data {
		val, ok := ds2.Get(context.Background(), k)
		assert.True(t, ok)
		assert.Equal(t, v, val)
	}
	for k, v := range ds.NodeData {
		for k1, v1 := range v {
			val, ok := ds2.GetDependNodeValue(context.Background(), k, k1)
			assert.True(t, ok)
			assert.Equal(t, v1, val)
		}
	}
}
