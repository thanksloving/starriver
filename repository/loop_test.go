package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/helper"
	"github.com/thanksloving/starriver/internal/core"
)

func TestLoopComponent_Execute(t *testing.T) {
	registerTestNode()
	registerTemplate()
	lc := &loopComponent{
		SkeletonWithParameter: helper.NewSkeletonWithParameter("test_loop", &loopParam{}),
	}

	dc := core.NewDataContext(context.Background(), &mockPipeline{}, nil)

	subConf := starriver.PipelineConf{
		Name: "test_inner_loop",
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
							Literal: "hello loop",
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

	param := &loopParam{
		Items:        []interface{}{1, 2, 3},
		PipelineConf: subConf,
		ItemKey:      "loop_item",
		IndexKey:     "loop_index",
	}

	resp := lc.Execute(dc, param)
	assert.True(t, resp.IsPass())
	assert.NotNil(t, resp.GetData()["Results"])

	results := resp.GetData()["Results"].([]map[string]interface{})
	assert.Equal(t, 3, len(results))
	for _, res := range results {
		assert.Equal(t, "hello loop", res["test_val"])
	}
}
