package dag

import (
	"context"
	"time"

	"github.com/thanksloving/starriver"
)

type nodeDataContext struct {
	PrevTaskIDList []string
	Properties     map[string]interface{}
	ctx            context.Context
	starriver.DataContext
}

func newNodeDataContext(dataContext starriver.DataContext, timeout *time.Duration) (newContext *nodeDataContext, cancel context.CancelFunc) {
	ctx := dataContext.Context()
	if timeout != nil {
		ctx, cancel = context.WithTimeout(dataContext.Context(), *timeout)
	}
	return &nodeDataContext{
		ctx:         ctx,
		DataContext: dataContext,
		Properties:  make(map[string]interface{}),
	}, cancel
}

func (ndc *nodeDataContext) AppendProperties(properties map[string]interface{}) {
	for k, v := range properties {
		ndc.Properties[k] = v
	}
}

func (ndc *nodeDataContext) AppendPrevTask(taskID string) {
	ndc.PrevTaskIDList = append(ndc.PrevTaskIDList, taskID)
}

// Get get value by key, first it will find it in edge's properties, then depend node's output, and finally in the shared data store.
func (ndc *nodeDataContext) Get(key string) (interface{}, bool) {
	if val, ok := ndc.Properties[key]; ok {
		return val, ok
	}
	for _, dependTaskID := range ndc.PrevTaskIDList {
		if val, ok := ndc.GetDependNodeValue(dependTaskID, key); ok {
			return val, ok
		}
	}
	return ndc.DataContext.Get(key)
}

func (ndc *nodeDataContext) Context() context.Context {
	return ndc.ctx
}

func (ndc *nodeDataContext) Deadline() (deadline time.Time, ok bool) {
	return ndc.ctx.Deadline()
}

func (ndc *nodeDataContext) Done() <-chan struct{} {
	return ndc.ctx.Done()
}

func (ndc *nodeDataContext) Err() error {
	return ndc.ctx.Err()
}

func (ndc *nodeDataContext) Value(key any) any {
	return ndc.ctx.Value(key)
}
