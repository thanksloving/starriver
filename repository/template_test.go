package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/helper"
	"github.com/thanksloving/starriver/internal/core"
)

type mockPipeline struct {
	env map[string]interface{}
}

func (p *mockPipeline) GetName() string { return "mock_pipeline" }
func (p *mockPipeline) GetTaskConfigure(taskId string) starriver.TaskConfigure {
	return starriver.TaskConfigure{}
}
func (p *mockPipeline) Run(dataContext starriver.DataContext) starriver.Result {
	return starriver.Result{}
}
func (p *mockPipeline) GetStatus() starriver.PipelineStatus { return starriver.PipelineStatusInit }
func (p *mockPipeline) GetTaskStatus(taskID string) starriver.TaskStatus {
	return starriver.TaskStatusInit
}
func (p *mockPipeline) SetTaskStatus(taskID string, state starriver.TaskStatus) bool {
	return true
}
func (p *mockPipeline) Env(key string) (interface{}, bool) {
	if p.env == nil {
		return nil, false
	}
	val, ok := p.env[key]
	return val, ok
}

func TestTemplateComponent_InvalidTemplate(t *testing.T) {
	tc := &templateComponent{
		SkeletonWithParameter: helper.NewSkeletonWithParameter("test_template", &templateParam{}),
	}

	p := &mockPipeline{}
	dc := core.NewDataContext(context.Background(), p, nil)

	// An invalid template syntax that would normally cause panic with template.Must
	invalidTmpl := "{{ .Invalid Syntax }}"
	param := &templateParam{
		Template:  invalidTmpl,
		OutputKey: "out",
	}

	// It should gracefully return an error response, not panic
	resp := tc.Execute(dc, param)
	assert.False(t, resp.IsPass())
	assert.Error(t, resp.GetError())
}

// Add embedded struct manually for the test just to satisfy interface or composition
// Alternatively, since helper.SkeletonWithParameter is an embedded field, we must initialize it properly
