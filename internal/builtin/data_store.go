package builtin

import (
	"context"
	"sync"

	"github.com/thanksloving/starriver"
)

// defaultSharedDataStore default Data store use memory to store the Data
type (
	defaultSharedDataStore struct {
		lock     sync.RWMutex `json:"-" yaml:"-" msgpack:"-"`
		nodeLock sync.RWMutex `json:"-" yaml:"-"  msgpack:"-"`
		Codec    Codec        `json:"-" yaml:"-" msgpack:"-"`
		dataStore
	}

	dataStore struct {
		Data     map[string]interface{}            `json:"data" yaml:"data" msgpack:"data"`
		NodeData map[string]map[string]interface{} `json:"node_data" yaml:"node_data" msgpack:"node_data"`
	}

	DataSourceOption func(*defaultSharedDataStore)
)

func NewSharedDataStore(options ...DataSourceOption) starriver.SharedDataStore {
	ds := &defaultSharedDataStore{
		dataStore: dataStore{
			NodeData: make(map[string]map[string]interface{}),
			Data:     make(map[string]interface{}),
		},
	}
	for _, option := range options {
		option(ds)
	}
	if ds.Codec == nil {
		ds.Codec = &defaultCodec{}
	}
	return ds
}

func WithCodec(codec Codec) DataSourceOption {
	return func(ds *defaultSharedDataStore) {
		ds.Codec = codec
	}
}

var _ starriver.SharedDataStore = (*defaultSharedDataStore)(nil)

func (sds *defaultSharedDataStore) Configure(_, _ string) {}

func (sds *defaultSharedDataStore) Get(_ context.Context, key string) (interface{}, bool) {
	sds.lock.RLock()
	defer sds.lock.RUnlock()
	res, ok := sds.Data[key]
	return res, ok
}

func (sds *defaultSharedDataStore) Set(_ context.Context, key string, value interface{}) bool {
	sds.lock.Lock()
	defer sds.lock.Unlock()
	_, ok := sds.Data[key]
	sds.Data[key] = value
	return ok
}

func (sds *defaultSharedDataStore) Del(_ context.Context, key string) {
	sds.lock.Lock()
	defer sds.lock.Unlock()
	delete(sds.Data, key)
}

func (sds *defaultSharedDataStore) SetCurrentNodeData(_ context.Context, nodeId string, data map[string]interface{}) {
	sds.nodeLock.Lock()
	defer sds.nodeLock.Unlock()
	sds.NodeData[nodeId] = data
}

func (sds *defaultSharedDataStore) GetDependNodeValue(_ context.Context, nodeId, key string) (interface{}, bool) {
	sds.nodeLock.RLock()
	defer sds.nodeLock.RUnlock()
	if data := sds.NodeData[nodeId]; len(data) > 0 {
		val, ok := data[key]
		return val, ok
	}
	return nil, false
}

func (sds *defaultSharedDataStore) Marshal() ([]byte, error) {
	sds.lock.RLock()
	sds.nodeLock.RLock()
	defer func() {
		sds.lock.RUnlock()
		sds.nodeLock.RUnlock()
	}()
	return sds.Codec.Marshal(sds.dataStore)
}

func (sds *defaultSharedDataStore) Unmarshal(data []byte) error {
	sds.lock.Lock()
	sds.nodeLock.Lock()
	defer func() {
		sds.lock.Unlock()
		defer sds.nodeLock.Unlock()
	}()
	var ds dataStore
	if err := sds.Codec.Unmarshal(data, &ds); err != nil {
		return err
	}
	sds.dataStore = ds
	return nil
}
