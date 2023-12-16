package flow

import (
	"context"
	"encoding/json"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/internal/builtin"
	"github.com/thanksloving/starriver/internal/core"
	"github.com/thanksloving/starriver/internal/util"
	"github.com/thanksloving/starriver/registry"
	_ "github.com/thanksloving/starriver/repository" // 注册系统内置的组件
)

type (
	RiverEngine struct {
		WorkerConcurrency int
		Semaphore         util.Semaphore
		LoggingEnabled    bool
		DebugEnabled      bool
		EventHandler      starriver.EventHandler
		cronClient        *cron.Cron
	}

	Option func(*RiverEngine)
)

var (
	NewDataContext = core.NewDataContext
	SetRequestID   = core.SetRequestID
	// NewSharedDataStore 新建自己的共享数据存储，主要用于修改默认的序列化方式。默认使用的 json 序列化方式对数字类型会有精度损失或类型错乱。
	NewSharedDataStore = builtin.NewSharedDataStore
	SetSharedDataStore = core.SetSharedDataStore
	SetLogger          = core.SetLogger
	SetLogLevel        = core.SetLogLevel
)

func LoadPipelineByYaml(yamlConf string) (*starriver.PipelineConf, error) {
	var pc starriver.PipelineConf
	if err := yaml.Unmarshal([]byte(yamlConf), &pc); err != nil {
		return nil, err
	}
	return &pc, nil
}

func LoadPipelineByJson(jsonConf string) (*starriver.PipelineConf, error) {
	var pc starriver.PipelineConf
	if err := json.Unmarshal([]byte(jsonConf), &pc); err != nil {
		return nil, err
	}
	return &pc, nil
}

func NewPipeline(conf starriver.PipelineConf) (starriver.Pipeline, error) {
	return core.BuildPipeline(conf, starriver.PipelineStatusInit, make(map[string]starriver.TaskStatus))
}

// Rebuild  a pipeline from a snapshot
func Rebuild(ctx context.Context, conf starriver.PipelineConf, taskStatuses map[string]starriver.TaskStatus,
	snapshot starriver.SharedDataStore, initialData map[string]interface{},
	opts ...core.ContextOption) (starriver.DataContext, starriver.Pipeline, error) {
	pipeline, err := core.BuildPipeline(conf, starriver.PipelineStatusBlocked, taskStatuses)
	if err != nil {
		return nil, nil, err
	}
	if snapshot != nil {
		opts = append(opts, core.SetSharedDataStore(snapshot))
	}
	dataContext := NewDataContext(ctx, pipeline, initialData, opts...)
	return dataContext, pipeline, nil
}

// NewRiverEngine new a river engine
func NewRiverEngine(options ...Option) *RiverEngine {
	re := &RiverEngine{
		WorkerConcurrency: 200,
		cronClient:        cron.New(cron.WithSeconds()),
	}
	for _, option := range options {
		option(re)
	}
	re.Semaphore = util.NewSemaphore(re.WorkerConcurrency)
	return re
}

func SetEventHandler(eventHandler starriver.EventHandler) Option {
	return func(re *RiverEngine) {
		re.EventHandler = eventHandler
	}
}

func GetComponents() []*starriver.Component {
	return registry.GetAllComponents()
}

func (re *RiverEngine) CronRun(spec string, pipeline starriver.Pipeline, data map[string]interface{}) {
	entryID, err := re.cronClient.AddFunc(spec, func() {
		dataContext := NewDataContext(context.Background(), pipeline, data)
		result := re.Run(dataContext, pipeline)
		dataContext.Infof("[Cron]spec=%q, pipeline=%q, data=%+v, result=%+v", spec, pipeline.GetName(), data, result)
	})
	logrus.Infof("[Cron]AddFunc, spec=%q, pipeline=%q, entryID=%q, err=%v", spec, pipeline.GetName(), entryID, err)
	re.cronClient.Start()
}

func (re *RiverEngine) Run(dataContext starriver.DataContext, pipeline starriver.Pipeline) starriver.Result {
	defer func() {
		if re.EventHandler != nil {
			switch pipeline.GetStatus() {
			case starriver.PipelineStatusBlocked:
				re.EventHandler.OnBlocked(dataContext)
			default:
				re.EventHandler.OnEnd(dataContext)
			}
		}
	}()
	if re.EventHandler != nil {
		switch pipeline.GetStatus() {
		case starriver.PipelineStatusBlocked:
			re.EventHandler.OnResume(dataContext)
		default:
			re.EventHandler.OnStart(dataContext)
		}
	}
	re.Semaphore.Acquire()
	defer re.Semaphore.Release()
	return pipeline.Run(dataContext)
}

func (re *RiverEngine) Destroy() {
	re.cronClient.Stop()
}
