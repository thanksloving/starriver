package core

import (
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/helper"
	"github.com/thanksloving/starriver/internal/dag"
	"github.com/thanksloving/starriver/internal/util"
)

type GraphWalker struct {
	ParallelSem util.Semaphore
	lock        sync.Locker
	serial      bool // execute the pipeline by serial, default is false
	Pipeline    starriver.Pipeline
}

func (walker *GraphWalker) callback(dataContext starriver.DataContext, vertex dag.Vertex) (resp starriver.Response) {
	switch walker.Pipeline.GetTaskStatus(vertex.ID()) {
	case starriver.TaskStatusSuccess, starriver.TaskStatusSkipped:
		return helper.NewSuccessResponse()
	case starriver.TaskStatusFailure:
		return helper.NewErrorResponse(nil)
	case starriver.TaskStatusInit:
		// 第一次跑
	case starriver.TaskStatusBlocked:
		// 重跑
	}
	ch := make(chan starriver.Response, 1)
	defer func() {
		close(ch)
		if resp == nil {
			helper.NewErrorResponse(fmt.Errorf("%q response is nil, unkonwn exeception", vertex.ID()))
			return
		}
	}()
	if walker.serial {
		walker.lock.Lock()
		defer walker.lock.Unlock()
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resp = helper.NewErrorResponse(fmt.Errorf("%v", r))
			}
		}()
		if executable, ok := vertex.(starriver.Executable); ok {
			resp = walker.execute(dataContext, executable)
		} else if _, ok := vertex.(dag.Node); ok {
			dataContext.Debugf("AnyNode Pass, %q", vertex.ID())
			resp = helper.NewSuccessResponse()
		} else {
			resp = helper.NewFatalResponse(fmt.Errorf("vertex do not support, %T", vertex))
		}
		ch <- resp
	}()
	select {
	case <-dataContext.Done():
		return helper.NewErrorResponse(dataContext.Err())
	case resp = <-ch:
		return resp
	}
}

func (walker *GraphWalker) Walk(graph dag.DAG, dataContext starriver.DataContext) error {
	responses := graph.Walk(dataContext, walker.callback)
	var errs *multierror.Error
	for _, resp := range responses {
		if resp.GetFailureLevel() > starriver.FailureLevelWarning && !resp.IsPass() {
			if err := resp.GetError(); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}
	return errs.ErrorOrNil()
}

func (walker *GraphWalker) execute(dataContext starriver.DataContext, executable starriver.Executable) (resp starriver.Response) {
	tc := walker.Pipeline.GetTaskConfigure(executable.ID())
	if tc.SkipExecution {
		dataContext.Pipeline().SetTaskStatus(executable.ID(), starriver.TaskStatusSkipped)
		return helper.NewSuccessResponse()
	}
	walker.ParallelSem.Acquire()
	defer func() {
		if r := recover(); r != nil {
			if resp == nil {
				resp = helper.NewErrorResponse(fmt.Errorf("%q executable execute panic, %v", executable.ID(), r))
			}
			dataContext.Errorf("component %q execute error, err=%v", executable.ID(), r)
		} else if ae, ok := executable.(starriver.AfterExecute); ok {
			ae.After(dataContext, resp)
		}
		walker.ParallelSem.Release()
	}()
	var param interface{}
	if p, ok := executable.(starriver.WithParameters); ok {
		var err error
		ap := &assembleParam{id: executable.ID()}
		if param, err = ap.prepareParameter(dataContext, tc.Params, p.ParameterNew()); err != nil {
			return helper.NewErrorResponse(err)
		}
	}
	if be, ok := executable.(starriver.BeforeExecute); ok {
		be.Before(dataContext)
	}
	resp = executable.Execute(dataContext, param)
	if tc.AbortIfError && resp.GetFailureLevel() > starriver.FailureLevelWarning {
		resp.SetFailureLevel(starriver.FailureLevelFatal)
	}
	if tc.AlwaysPass && !resp.IsPass() {
		resp.SetPass(true)
	}
	dataContext.Debugf("component %q execute done, resp=%+v", executable.ID(), resp)
	if listener, ok := executable.(starriver.Listener); ok {
		if resp.GetFailureLevel() == starriver.FailureLevelNormal {
			listener.OnSuccess(dataContext, resp.GetData())
		} else {
			listener.OnFailure(dataContext, resp.GetError())
		}
	}
	return resp
}
