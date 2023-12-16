package core

import (
	"sync"
	"time"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/internal/dag"
)

var (
	_ starriver.Pipeline = (*pipeline)(nil)
)

type (
	pipeline struct {
		Name           string
		env            map[string]interface{}
		status         starriver.PipelineStatus
		walker         GraphWalker
		TaskConfigures map[string]starriver.TaskConfigure
		lock           sync.RWMutex
		TaskStatuses   map[string]starriver.TaskStatus
		ResultKeys     []string
		Timeout        *time.Duration
		Graph          dag.DAG
	}
)

func (p *pipeline) GetName() string {
	return p.Name
}

func (p *pipeline) GetTaskConfigure(taskId string) starriver.TaskConfigure {
	if config, ok := p.TaskConfigures[taskId]; ok {
		return config
	}
	return starriver.TaskConfigure{}
}

func (p *pipeline) GetStatus() starriver.PipelineStatus {
	return p.status
}

func (p *pipeline) Env(key string) (interface{}, bool) {
	if p.env == nil {
		return nil, false
	}
	val, ok := p.env[key]
	return val, ok
}

func (p *pipeline) GetTaskStatus(taskID string) starriver.TaskStatus {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.TaskStatuses[taskID]
}

func (p *pipeline) SetTaskStatus(taskID string, state starriver.TaskStatus) bool {
	p.lock.Lock()
	defer p.lock.Unlock()
	switch p.TaskStatuses[taskID] {
	case starriver.TaskStatusSuccess, starriver.TaskStatusSkipped, starriver.TaskStatusFailure:
		return false
	case starriver.TaskStatusBlocked, starriver.TaskStatusInit:
		p.TaskStatuses[taskID] = state
	}
	return true
}

func (p *pipeline) checkBlocked(dataContext starriver.DataContext) *starriver.Result {
	for _, taskStatus := range p.TaskStatuses {
		if taskStatus != starriver.TaskStatusBlocked {
			continue
		}
		p.status = starriver.PipelineStatusBlocked
		snapshot, err := dataContext.Marshal()
		if err != nil {
			dataContext.Errorf("[pipeline]%q snapshot error %v", p.Name, err)
		}
		return &starriver.Result{
			Status:   p.status,
			State:    p.TaskStatuses,
			Snapshot: snapshot,
		}
	}
	return nil
}

func (p *pipeline) assembleResult(dataContext starriver.DataContext) map[string]interface{} {
	if len(p.ResultKeys) == 0 {
		return nil
	}
	data := make(map[string]interface{})
	leaves, _ := p.Graph.Leaves()
	getResultFunc := func(key string) (interface{}, bool) {
		// 获得叶子节点的结果，优先看看是不是在叶子节点里
		for _, leaf := range leaves {
			if val, ok := dataContext.GetDependNodeValue(leaf.ID(), key); ok {
				return val, true
			}
		}
		if val, ok := dataContext.Get(key); ok {
			return val, true
		}
		return nil, false
	}
	for _, key := range p.ResultKeys {
		if val, ok := getResultFunc(key); ok {
			data[key] = val
		} else {
			dataContext.Errorf("[pipeline]%q result key %q not exist", p.Name, key)
		}
	}
	return data
}

func (p *pipeline) Run(dataContext starriver.DataContext) starriver.Result {
	defer dataContext.Release()
	if p.Timeout != nil {
		cancel := dataContext.WithTimeout(*p.Timeout)
		defer cancel()
	}
	if err := p.walker.Walk(p.Graph, dataContext); err != nil {
		p.status = starriver.PipelineStatusFailure
		return starriver.Result{
			Status: p.status,
			State:  p.TaskStatuses,
			Error:  err,
		}
	}
	if result := p.checkBlocked(dataContext); result != nil {
		return *result
	}
	p.status = starriver.PipelineStatusSuccess
	data := p.assembleResult(dataContext)
	return starriver.Result{
		Data:   data,
		Status: p.status,
		State:  p.TaskStatuses,
	}
}
