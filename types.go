package starriver

import (
	"reflect"
	"time"
)

const (
	BuiltinNodePrefix = "@"

	PipelineStatusInit    PipelineStatus = "init"
	PipelineStatusBlocked PipelineStatus = "blocked" // 需要人工介入
	PipelineStatusFailure PipelineStatus = "failure"
	PipelineStatusSuccess PipelineStatus = "success"

	TaskStatusInit    TaskStatus = "init"
	TaskStatusBlocked TaskStatus = "blocked"
	TaskStatusSkipped TaskStatus = "skipped"
	TaskStatusSuccess TaskStatus = "success"
	TaskStatusFailure TaskStatus = "failure"

	ParamTypeLiteral  ParamType = "literal"
	ParamTypeVariable ParamType = "variable"
	ParamTypeComplex  ParamType = "complex"
	ParamTypeMapping  ParamType = "mapping"

	ConditionGT ConditionOperator = ">"
	ConditionLT ConditionOperator = "<"
	ConditionLE ConditionOperator = "<="
	ConditionGE ConditionOperator = ">="
	ConditionEQ ConditionOperator = "=="
	ConditionNE ConditionOperator = "!="
	ConditionIn ConditionOperator = "in"
)

const (
	FailureLevelNormal FailureLevel = iota
	FailureLevelWarning
	FailureLevelError
	FailureLevelFatal
)

type (
	TaskStatus     string
	PipelineStatus string

	ConditionOperator string

	GraphObject interface {
		ID() string
	}

	Node = GraphObject

	TaskConfigure struct {
		Timeout       *time.Duration `json:"timeout" yaml:"timeout"`
		AlwaysPass    bool           `json:"always_pass" yaml:"always_pass"`       // the node's result will always pass even error
		SkipExecution bool           `json:"skip_execution" yaml:"skip_execution"` // skip the executor
		AbortIfError  bool           `json:"abort_if_error" yaml:"abort_if_error"` // abort the pipeline when error
		Params        Params         `json:"params" yaml:"params"`                 // custom params
	}

	Params []Param

	ParamType string

	Param struct {
		Name     string           `json:"name" yaml:"name"` // the struct filed name, required
		Type     ParamType        `json:"type" yaml:"type"`
		Variable string           `json:"variable" yaml:"variable"` // required when ParamTypeVariable,
		Literal  interface{}      `json:"literal" yaml:"literal"`   // required when ParamTypeLiteral
		Complex  Params           `json:"complex" yaml:"complex"`   // required when ParamTypeComplex, the value is an param slice
		Mapping  map[string]Param `json:"mapping" yaml:"mapping"`   // required when ParamTypeMapping
		Required bool             `json:"required" yaml:"required"` // error or ignore when missing
	}

	Pipeline interface {
		GetName() string
		GetTaskConfigure(taskId string) TaskConfigure
		Run(dataContext DataContext) Result
		GetStatus() PipelineStatus
		GetTaskStatus(taskID string) TaskStatus
		SetTaskStatus(taskID string, state TaskStatus) bool
		Env(key string) (interface{}, bool)
	}

	Component struct {
		Name      string  `json:"name"`
		Namespace *string `json:"namespace"`
		Desc      string  `json:"desc"`
		Executor  func(id string) Executable
		Input     []InputParam           `json:"input"`
		Output    map[string]OutputValue `json:"output"`
		Timeout   *time.Duration         `json:"timeout"`
	}

	InputParam struct {
		Key      string
		Desc     string
		Required bool
		Type     reflect.Kind  // 输入参数类型
		Options  []interface{} // 可选项，如果是限制输入的，可以有可选项，下拉列表
	}

	OutputValue struct {
		Desc string
		Type reflect.Kind
	}

	BeforeExecute interface {
		Before(dataContext DataContext)
	}

	Executable interface {
		ID() string
		Execute(dataContext DataContext, param interface{}) Response
	}

	WithParameters interface {
		ParameterNew() interface{}
	}

	AfterExecute interface {
		After(dataContext DataContext, resp Response)
	}

	Listener interface {
		OnSuccess(dataContext DataContext, result map[string]interface{})
		OnFailure(dataContext DataContext, err error)
	}

	EventHandler interface {
		// OnStart start pipeline
		OnStart(dataContext DataContext)
		// OnEnd end pipeline
		OnEnd(dataContext DataContext)
		// OnResume resume pipeline
		OnResume(dataContext DataContext)
		// OnBlocked break pipeline
		OnBlocked(dataContext DataContext)
	}

	FailureLevel int64

	Result struct {
		Data     map[string]interface{}
		Snapshot []byte
		Status   PipelineStatus
		State    map[string]TaskStatus
		Error    error
	}

	Response interface {
		IsPass() bool
		GetData() map[string]interface{}
		GetFailureLevel() FailureLevel
		GetError() error
		GetStatus() TaskStatus
		SetPass(pass bool)
		SetFailureLevel(level FailureLevel)
	}

	Responses map[string]Response
)
