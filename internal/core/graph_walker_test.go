package core

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/helper"
	"github.com/thanksloving/starriver/internal/util"
)

type mockExecutable struct {
	id         string
	delay      time.Duration
	shouldFail bool
}

func (m *mockExecutable) ID() string {
	return m.id
}

func (m *mockExecutable) Execute(dataContext starriver.DataContext, param interface{}) starriver.Response {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.shouldFail {
		return helper.NewErrorResponse(fmt.Errorf("mock error"))
	}
	return helper.NewSuccessResponse()
}

type mockPipeline struct {
	status       starriver.PipelineStatus
	taskStatuses map[string]starriver.TaskStatus
	lock         sync.Mutex
}

func (p *mockPipeline) GetName() string { return "mock_pipeline" }
func (p *mockPipeline) GetTaskConfigure(taskId string) starriver.TaskConfigure {
	return starriver.TaskConfigure{}
}
func (p *mockPipeline) Run(dataContext starriver.DataContext) starriver.Result {
	return starriver.Result{}
}
func (p *mockPipeline) GetStatus() starriver.PipelineStatus { return p.status }
func (p *mockPipeline) GetTaskStatus(taskID string) starriver.TaskStatus {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.taskStatuses[taskID]
}
func (p *mockPipeline) SetTaskStatus(taskID string, state starriver.TaskStatus) bool {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.taskStatuses[taskID] = state
	return true
}
func (p *mockPipeline) Env(key string) (interface{}, bool) { return nil, false }

func TestGraphWalker_Callback_ContextDone(t *testing.T) {
	t.Run("context cancelled before completion", func(t *testing.T) {
		p := &mockPipeline{
			taskStatuses: map[string]starriver.TaskStatus{
				"task1": starriver.TaskStatusInit,
			},
		}
		walker := &GraphWalker{
			ParallelSem: util.NewSemaphore(10),
			Pipeline:    p,
		}

		ctx, cancel := context.WithCancel(context.Background())
		dc := NewDataContext(ctx, p, nil)

		// Create an executable that takes 100ms
		exec := &mockExecutable{id: "task1", delay: 100 * time.Millisecond}

		// Run callback in a goroutine
		var resp starriver.Response
		done := make(chan struct{})
		go func() {
			resp = walker.callback(dc, exec)
			close(done)
		}()

		// Cancel context immediately
		cancel()

		<-done
		assert.NotNil(t, resp)
		assert.False(t, resp.IsPass())
		assert.ErrorIs(t, resp.GetError(), context.Canceled)
		
		// Wait a bit to ensure the goroutine inside callback finishes and doesn't panic on closed channel
		time.Sleep(200 * time.Millisecond)
	})
}
