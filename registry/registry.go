package registry

import (
	"sync"
	"time"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/internal/dag"
)

type (
	ExecutableFunc func(id string) starriver.Executable

	nodeFunc func(string) starriver.Node

	registry struct {
		lock              sync.RWMutex
		builtinNodes      map[string]nodeFunc
		defaultComponents map[string]*starriver.Component
		customComponents  map[string]map[string]*starriver.Component
	}

	RegisterOption func(*starriver.Component)
)

var (
	instance *registry
)

func Namespace(namespace string) RegisterOption {
	return func(component *starriver.Component) {
		component.Namespace = &namespace
	}
}

func Timeout(timeout *time.Duration) RegisterOption {
	return func(component *starriver.Component) {
		component.Timeout = timeout
	}
}

func Input(input []starriver.InputParam) RegisterOption {
	return func(component *starriver.Component) {
		component.Input = input
	}
}

func Output(output map[string]starriver.OutputValue) RegisterOption {
	return func(component *starriver.Component) {
		component.Output = output
	}
}

func init() {
	instance = &registry{
		builtinNodes:      make(map[string]nodeFunc),
		defaultComponents: make(map[string]*starriver.Component),
		customComponents:  make(map[string]map[string]*starriver.Component),
	}
	for _, nt := range []dag.NodeType{dag.NodeTypeNot, dag.NodeTypeAny} {
		instance.builtinNodes[string(nt)] = func(nt dag.NodeType) nodeFunc {
			return func(id string) starriver.Node {
				return dag.NewNode(id, nt)
			}
		}(nt)
	}
}

func Register(name, desc string, fn ExecutableFunc, options ...RegisterOption) {
	component := &starriver.Component{
		Name:     name,
		Desc:     desc,
		Executor: fn,
	}
	for _, option := range options {
		option(component)
	}
	instance.lock.Lock()
	defer instance.lock.Unlock()
	if component.Namespace == nil {
		instance.defaultComponents[component.Name] = component
		return
	}
	if nodes, ok := instance.customComponents[*component.Namespace]; ok {
		nodes[component.Name] = component
	} else {
		instance.customComponents[*component.Namespace] = map[string]*starriver.Component{
			component.Name: component,
		}
	}
}

func GetAllComponents() []*starriver.Component {
	components := make([]*starriver.Component, 0)
	instance.lock.RLock()
	defer instance.lock.RUnlock()
	for _, component := range instance.defaultComponents {
		components = append(components, component)
	}
	for _, componentMap := range instance.customComponents {
		for _, component := range componentMap {
			components = append(components, component)
		}
	}
	return components
}

func GetComponent(name string, namespace *string) *starriver.Component {
	instance.lock.RLock()
	defer instance.lock.RUnlock()
	if namespace == nil {
		instance.lock.RLock()
		defer instance.lock.RUnlock()
		return instance.defaultComponents[name]
	}
	if components := instance.customComponents[*namespace]; components != nil {
		return components[name]
	}
	return nil
}

func NewBuiltinNode(id, name string) starriver.Node {
	if nodeFunc := instance.builtinNodes[name[1:]]; nodeFunc != nil {
		return nodeFunc(id)
	}
	return nil
}
