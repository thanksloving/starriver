package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/internal/builtin"
)

type (
	ContextOption func(*dataContext)

	dataContext struct {
		requestID string
		ctx       context.Context
		cancel    context.CancelFunc

		pipeline        starriver.Pipeline
		SharedDataStore starriver.SharedDataStore
		Logger          starriver.Logger
	}
)

var (
	contextPool = sync.Pool{
		New: func() interface{} {
			return new(dataContext)
		},
	}

	_ starriver.DataContext = (*dataContext)(nil)
)

func NewDataContext(ctx context.Context, pipeline starriver.Pipeline, initialData map[string]interface{},
	opts ...ContextOption) starriver.DataContext {
	if pipeline == nil {
		panic("pipeline is nil")
	}
	sc := contextPool.Get().(*dataContext)
	sc.ctx, sc.cancel = context.WithCancel(ctx)
	sc.Logger = builtin.NewLogger()
	sc.Logger.SetLoggerLevel(starriver.DebugLevel)
	for _, opt := range opts {
		opt(sc)
	}
	if sc.requestID == "" {
		if traceId := ctx.Value("X-B3-Traceid"); traceId != nil {
			sc.requestID = traceId.(string)
		} else {
			sc.requestID = uuid.New().String()
		}
	}
	if sc.SharedDataStore == nil {
		sc.SharedDataStore = builtin.NewSharedDataStore()
	}
	sc.pipeline = pipeline
	sc.requestID = fmt.Sprintf("%s#%s", pipeline.GetName(), sc.requestID)
	for k, v := range initialData {
		sc.SharedDataStore.Set(ctx, k, v)
	}
	return sc
}

func SetLogLevel(level starriver.LogLevel) ContextOption {
	return func(dc *dataContext) {
		dc.Logger.SetLoggerLevel(level)
	}
}

func SetRequestID(requestID string) ContextOption {
	return func(dc *dataContext) {
		dc.requestID = requestID
	}
}

func SetSharedDataStore(sharedDataStore starriver.SharedDataStore) ContextOption {
	return func(dc *dataContext) {
		dc.SharedDataStore = sharedDataStore
	}
}

func SetLogger(logger starriver.Logger) ContextOption {
	return func(dc *dataContext) {
		dc.Logger = logger
	}
}

func (dc *dataContext) Release() {
	dc.cancel()
	dc.ctx = nil
	dc.cancel = nil
	dc.requestID = ""
	dc.SharedDataStore = nil
	dc.Logger = nil
	dc.pipeline = nil
	contextPool.Put(dc)
}

func (dc *dataContext) Env(key string) (interface{}, bool) {
	return dc.pipeline.Env(key)
}

func (dc *dataContext) Context() context.Context {
	return dc.ctx
}

func (dc *dataContext) Pipeline() starriver.Pipeline {
	return dc.pipeline
}

func (dc *dataContext) GetRequestID() string {
	return dc.requestID
}

func (dc *dataContext) Deadline() (deadline time.Time, ok bool) {
	return dc.ctx.Deadline()
}

func (dc *dataContext) Done() <-chan struct{} {
	return dc.ctx.Done()
}

func (dc *dataContext) Err() error {
	return dc.ctx.Err()
}

func (dc *dataContext) Value(key any) any {
	return dc.ctx.Value(key)
}

func (dc *dataContext) WithValue(key, value any) {
	dc.ctx = context.WithValue(dc.ctx, key, value)
}

func (dc *dataContext) Stop() {
	dc.cancel()
}

func (dc *dataContext) WithTimeout(timeout time.Duration) context.CancelFunc {
	var cancel context.CancelFunc
	dc.ctx, cancel = context.WithTimeout(dc.ctx, timeout)
	return cancel
}
