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
	loopComponent struct {
		helper.SkeletonWithParameter
	}

	loopParam struct {
		Items        []interface{}          `json:"items"`
		PipelineConf starriver.PipelineConf `json:"pipeline_conf"`
		ItemKey      string                 `json:"item_key"`       // the key to store the current item in the sub-pipeline's data context
		IndexKey     string                 `json:"index_key"`      // the key to store the current index in the sub-pipeline's data context
		InputData    map[string]interface{} `json:"input_data"`
		MaxLoop      int                    `json:"max_loop"`       // safeguard against infinite loops if items is too large
		BreakOnError bool                   `json:"break_if_error"` // break loop if sub-pipeline fails
	}
)

var _ starriver.Executable = (*loopComponent)(nil)

func registerLoop() {
	registry.Register("Loop", "循环节点，对数组中的每个元素执行一个子流程",
		func(id string) starriver.Executable {
			return &loopComponent{helper.NewSkeletonWithParameter(id, &loopParam{})}
		},
		registry.Input([]starriver.InputParam{
			{
				Key:      "Items",
				Required: true,
				Desc:     "需要遍历的数组",
				Type:     reflect.Slice,
			},
			{
				Key:      "PipelineConf",
				Required: true,
				Desc:     "每次循环执行的子流程配置",
			},
			{
				Key:      "ItemKey",
				Required: false,
				Desc:     "传递当前元素到子流程的变量名（默认: loop_item）",
			},
			{
				Key:      "IndexKey",
				Required: false,
				Desc:     "传递当前索引到子流程的变量名（默认: loop_index）",
			},
			{
				Key:      "InputData",
				Required: false,
				Desc:     "传递给每次子流程的公共初始数据",
			},
			{
				Key:      "MaxLoop",
				Required: false,
				Desc:     "最大循环次数（默认: 不限制）",
			},
			{
				Key:      "BreakOnError",
				Required: false,
				Desc:     "如果子流程执行失败，是否中断循环（默认: true）",
			},
		}),
		registry.Output(map[string]starriver.OutputValue{
			"Results": {
				Desc: "所有子流程的执行结果数组",
				Type: reflect.Slice,
			},
		}),
	)
}

func (l *loopComponent) ParameterNew() interface{} {
	return &loopParam{
		ItemKey:      "loop_item",
		IndexKey:     "loop_index",
		BreakOnError: true,
	}
}

func (l *loopComponent) Execute(dataContext starriver.DataContext, param interface{}) starriver.Response {
	p := param.(*loopParam)
	
	if p.ItemKey == "" {
		p.ItemKey = "loop_item"
	}
	if p.IndexKey == "" {
		p.IndexKey = "loop_index"
	}

	results := make([]map[string]interface{}, 0, len(p.Items))
	
	for i, item := range p.Items {
		if p.MaxLoop > 0 && i >= p.MaxLoop {
			dataContext.Warnf("loop reached max_loop limit: %d", p.MaxLoop)
			break
		}

		select {
		case <-dataContext.Done():
			return helper.NewErrorResponse(dataContext.Err())
		default:
		}

		// Create a new input data map for this iteration, including the item and index
		iterData := make(map[string]interface{})
		for k, v := range p.InputData {
			iterData[k] = v
		}
		iterData[p.ItemKey] = item
		iterData[p.IndexKey] = i

		// Build and run the sub-pipeline for this iteration
		subPipeline, err := core.BuildPipeline(p.PipelineConf, starriver.PipelineStatusInit, make(map[string]starriver.TaskStatus))
		if err != nil {
			return helper.NewErrorResponse(fmt.Errorf("build loop sub pipeline error at index %d: %v", i, err))
		}

		ctx := context.WithValue(dataContext.Context(), "X-B3-Traceid", fmt.Sprintf("%s-loop-%d", dataContext.GetRequestID(), i))
		subDataContext := core.NewDataContext(ctx, subPipeline, iterData)
		
		result := subPipeline.Run(subDataContext)

		if result.Status != starriver.PipelineStatusSuccess {
			if p.BreakOnError {
				return helper.NewErrorResponse(fmt.Errorf("loop sub pipeline executed failed at index %d, err: %v", i, result.Error))
			} else {
				dataContext.Warnf("loop sub pipeline executed failed at index %d, err: %v, continuing...", i, result.Error)
				results = append(results, nil)
				continue
			}
		}

		results = append(results, result.Data)
	}

	return helper.NewSuccessDataResponse(map[string]interface{}{
		"Results": results,
	})
}
