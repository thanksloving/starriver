package repository

import (
	"context"
	"fmt"
	"reflect"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/helper"
	"github.com/thanksloving/starriver/internal/core"
	"github.com/thanksloving/starriver/registry"
)

type (
	subPipelineComponent struct {
		helper.SkeletonWithParameter
	}

	subPipelineParam struct {
		PipelineConf starriver.PipelineConf `json:"pipeline_conf"`
		InputData    map[string]interface{} `json:"input_data"`
	}
)

var _ starriver.Executable = (*subPipelineComponent)(nil)

func registerSubPipeline() {
	registry.Register("SubPipeline", "子流程节点，允许在当前流程中嵌入并执行另一个流程",
		func(id string) starriver.Executable {
			return &subPipelineComponent{helper.NewSkeletonWithParameter(id, &subPipelineParam{})}
		},
		registry.Input([]starriver.InputParam{
			{
				Key:      "PipelineConf",
				Required: true,
				Desc:     "子流程的配置 (starriver.PipelineConf)",
			},
			{
				Key:      "InputData",
				Required: false,
				Desc:     "传递给子流程的初始数据",
			},
		}),
		registry.Output(map[string]starriver.OutputValue{
			"Result": {
				Desc: "子流程的执行结果",
				Type: reflect.Map,
			},
		}),
	)
}

func (s *subPipelineComponent) ParameterNew() interface{} {
	return &subPipelineParam{}
}

func (s *subPipelineComponent) Execute(dataContext starriver.DataContext, param interface{}) starriver.Response {
	p := param.(*subPipelineParam)
	
	// Create sub-pipeline
	subPipeline, err := core.BuildPipeline(p.PipelineConf, starriver.PipelineStatusInit, make(map[string]starriver.TaskStatus))
	if err != nil {
		return helper.NewErrorResponse(fmt.Errorf("build sub pipeline error: %v", err))
	}
	
	// Create context for sub-pipeline
	// We inherit the request ID and timeout if possible, but use a new data store
	ctx := context.WithValue(dataContext.Context(), "X-B3-Traceid", dataContext.GetRequestID())
	subDataContext := core.NewDataContext(ctx, subPipeline, p.InputData)
	
	// Run sub-pipeline
	result := subPipeline.Run(subDataContext)
	
	if result.Status != starriver.PipelineStatusSuccess {
		return helper.NewErrorResponse(fmt.Errorf("sub pipeline executed failed with status: %s, err: %v", result.Status, result.Error))
	}
	
	return helper.NewSuccessDataResponse(map[string]interface{}{
		"Result": result.Data,
	})
}
