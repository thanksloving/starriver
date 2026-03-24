package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/helper"
	"github.com/thanksloving/starriver/internal/core"
)

func TestSubPipelineComponent_Execute(t *testing.T) {
	registerTestNode()
	registerTemplate()
	sp := &subPipelineComponent{
		SkeletonWithParameter: helper.NewSkeletonWithParameter("test_sub", &subPipelineParam{}),
	}

	dc := core.NewDataContext(context.Background(), &mockPipeline{}, nil)

	subConf := starriver.PipelineConf{
		Name: "test_inner",
		Result: []string{"test_val"},
		Pipeline: []starriver.Task{
			{
				ID:   "task1",
				Name: "TestNode",
				Config: starriver.TaskConfigure{
					Params: []starriver.Param{
						{
							Name: "Pass",
							Type: starriver.ParamTypeLiteral,
							Literal: true,
						},
					},
				},
			},
			{
				ID:   "task2",
				Name: "Template",
				Config: starriver.TaskConfigure{
					Params: []starriver.Param{
						{
							Name: "Template",
							Type: starriver.ParamTypeLiteral,
							Literal: "hello world",
						},
						{
							Name: "OutputKey",
							Type: starriver.ParamTypeLiteral,
							Literal: "test_val",
						},
						{
							Name: "Shared",
							Type: starriver.ParamTypeLiteral,
							Literal: true,
						},
					},
				},
				Depends: []starriver.Depend{
					{ID: "task1"},
				},
			},
		},
	}

	param := &subPipelineParam{
		PipelineConf: subConf,
	}

	resp := sp.Execute(dc, param)
	assert.True(t, resp.IsPass())
	assert.NotNil(t, resp.GetData()["Result"])
	
	resultMap := resp.GetData()["Result"].(map[string]interface{})
	assert.Equal(t, "hello world", resultMap["test_val"])
}
