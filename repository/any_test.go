package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/flow"
)

func TestAnyNode(t *testing.T) {
	conf := starriver.PipelineConf{
		Name: "any_node_test",
		Pipeline: []starriver.Task{
			{
				ID:   "task1",
				Name: "TestNode",
				Config: starriver.TaskConfigure{
					Params: []starriver.Param{
						{
							Name:    "Pass",
							Type:    starriver.ParamTypeLiteral,
							Literal: true,
						},
					},
				},
			},
			{
				ID:   "task2",
				Name: "TestNode",
				Config: starriver.TaskConfigure{
					Params: []starriver.Param{
						{
							Name:    "Pass",
							Type:    starriver.ParamTypeLiteral,
							Literal: true,
						},
					},
				},
				Depends: []starriver.Depend{
					{
						ID: "task1",
					},
				},
			},
			{
				ID:   "task3",
				Name: "TestNode",
				Config: starriver.TaskConfigure{
					Params: []starriver.Param{
						{
							Name:    "Pass",
							Type:    starriver.ParamTypeLiteral,
							Literal: false,
						},
					},
				},
				Depends: []starriver.Depend{
					{
						ID: "task1",
					},
				},
			},
			{
				ID:   "any",
				Name: "@any",
				Depends: []starriver.Depend{
					{
						ID: "task2",
					},
					{
						ID: "task3",
					},
				},
			},
			{
				ID:   "task4",
				Name: "TestNode",
				Config: starriver.TaskConfigure{
					Params: []starriver.Param{
						{
							Name:    "Pass",
							Type:    starriver.ParamTypeLiteral,
							Literal: true,
						},
					},
				},
				Depends: []starriver.Depend{
					{
						ID: "any",
					},
				},
			},
		},
	}
	pipeline, err := flow.NewPipeline(conf)
	assert.NoError(t, err)
	dc := flow.NewDataContext(context.Background(), pipeline, nil)
	engine := flow.NewRiverEngine()
	defer engine.Destroy()
	result := engine.Run(dc, pipeline)
	assert.Equal(t, result.Status, starriver.PipelineStatusSuccess)
	assert.Equal(t, result.State["task1"], starriver.TaskStatusSuccess)
	assert.Equal(t, result.State["task2"], starriver.TaskStatusSuccess)
	assert.Equal(t, result.State["task4"], starriver.TaskStatusSuccess)
	assert.Equal(t, result.State["task3"], starriver.TaskStatusFailure)
	assert.Equal(t, result.State["any"], starriver.TaskStatusSuccess)
}

func TestNotNode(t *testing.T) {
	conf := starriver.PipelineConf{
		Name: "not_node_test",
		Pipeline: []starriver.Task{
			{
				ID:   "task1",
				Name: "TestNode",
				Config: starriver.TaskConfigure{
					Params: []starriver.Param{
						{
							Name:    "Pass",
							Type:    starriver.ParamTypeLiteral,
							Literal: false,
						},
					},
				},
			},
			{
				ID:   "task2",
				Name: "@not",
				Depends: []starriver.Depend{
					{
						ID: "task1",
					},
				},
			},
			{
				ID:   "task3",
				Name: "TestNode",
				Config: starriver.TaskConfigure{
					Params: []starriver.Param{
						{
							Name:    "Pass",
							Type:    starriver.ParamTypeLiteral,
							Literal: true,
						},
					},
				},
				Depends: []starriver.Depend{
					{
						ID: "task2",
					},
				},
			},
		},
	}
	pipeline, err := flow.NewPipeline(conf)
	assert.NoError(t, err)
	dc := flow.NewDataContext(context.Background(), pipeline, nil)
	engine := flow.NewRiverEngine()
	defer engine.Destroy()
	result := engine.Run(dc, pipeline)
	assert.Equal(t, result.Status, starriver.PipelineStatusSuccess)
	assert.Equal(t, result.State["task1"], starriver.TaskStatusFailure)
	assert.Equal(t, result.State["task2"], starriver.TaskStatusSuccess)
	assert.Equal(t, result.State["task3"], starriver.TaskStatusSuccess)
}
