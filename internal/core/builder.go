package core

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/internal/dag"
	"github.com/thanksloving/starriver/internal/util"
	"github.com/thanksloving/starriver/registry"
)

func BuildPipeline(pc starriver.PipelineConf, status starriver.PipelineStatus, taskStatuses map[string]starriver.TaskStatus) (_ starriver.Pipeline, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("build pipeline %s panic: %v", pc.Name, r)
			logrus.Errorf("build pipeline %s panic: %v", pc.Name, r)
		}
	}()
	pipeline := &pipeline{
		env:          pc.Env,
		Name:         pc.Name,
		status:       status,
		ResultKeys:   pc.Result,
		Timeout:      pc.Timeout,
		TaskStatuses: taskStatuses,
	}
	tc := make(map[string]starriver.TaskConfigure)
	nodes := make(map[string]dag.Vertex)
	graph := dag.Graph{}
	for _, task := range pc.Pipeline {
		tc[task.ID] = task.Config
		var node starriver.Node
		if strings.HasPrefix(task.Name, starriver.BuiltinNodePrefix) {
			node = registry.NewBuiltinNode(task.ID, task.Name)
		} else {
			component := registry.GetComponent(task.Name, task.Namespace)
			if component == nil {
				return nil, fmt.Errorf("can not found node with name %q", task.Name)
			}
			node = component.Executor(task.ID)
			if tc[task.ID].Timeout == nil && component.Timeout != nil {
				*tc[task.ID].Timeout = *component.Timeout
			}
		}
		nodes[node.ID()] = node
		graph.Add(node)
		if _, ok := pipeline.TaskStatuses[task.ID]; !ok {
			pipeline.TaskStatuses[task.ID] = starriver.TaskStatusInit
		}
	}
	pipeline.TaskConfigures = tc
	for _, task := range pc.Pipeline {
		target := nodes[task.ID]
		for _, depend := range task.Depends {
			source := nodes[depend.ID]
			if depend.Condition != nil {
				graph.Connect(dag.
					ConditionEdge(source, target,
						depend.Condition.Key, depend.Condition.Value, depend.Condition.Operator).
					WithProperties(depend.Properties))
			} else {
				graph.Connect(dag.PropertyEdge(source, target, depend.Properties))
			}
		}
	}
	acyclicGraph := dag.NewDAG(graph)
	if err := acyclicGraph.Validate(); err != nil {
		return nil, err
	}
	sem := 10
	if pc.Concurrency != nil {
		sem = *pc.Concurrency
	}
	pipeline.Graph = acyclicGraph
	pipeline.walker = GraphWalker{
		ParallelSem: util.NewSemaphore(sem),
		Pipeline:    pipeline,
	}
	return pipeline, nil
}
