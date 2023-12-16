package starriver

import (
	"context"
	"time"
)

type (
	DataContext interface {
		Env(string) (interface{}, bool)
		Context() context.Context
		GetRequestID() string
		Pipeline() Pipeline
		Release()

		dataContextLogger
		dataContextSharedDataStore

		Stop()
		context.Context
		WithValue(key, value any)
		WithTimeout(duration time.Duration) context.CancelFunc
	}

	dataContextSharedDataStore interface {
		Set(key string, val interface{}) bool
		Del(key string)
		Get(key string) (interface{}, bool)

		Marshal() ([]byte, error)

		Configure(flowName, requestID string)

		SetCurrentNodeData(nodeId string, data map[string]interface{})
		GetDependNodeValue(nodeId, key string) (interface{}, bool)
	}

	SharedDataStore interface {
		Set(ctx context.Context, key string, val interface{}) bool
		Del(ctx context.Context, key string)
		Get(ctx context.Context, key string) (interface{}, bool)

		Marshal() ([]byte, error)
		Unmarshal(bs []byte) error

		Configure(flowName, requestID string)

		SetCurrentNodeData(ctx context.Context, nodeId string, data map[string]interface{})
		GetDependNodeValue(ctx context.Context, nodeId, key string) (interface{}, bool)
	}
)
