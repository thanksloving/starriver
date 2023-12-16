package starriver

import "time"

type (
	PipelineConf struct {
		Name        string                 `yaml:"name" json:"name"`
		Concurrency *int                   `yaml:"concurrency" json:"concurrency"`
		Result      []string               `yaml:"result" json:"result"`
		Timeout     *time.Duration         `yaml:"timeout" json:"timeout"`
		Env         map[string]interface{} `yaml:"env" json:"env"`
		Pipeline    []Task                 `yaml:"pipeline" json:"pipeline"`
	}

	Task struct {
		ID        string        `yaml:"task" json:"task"`
		Name      string        `yaml:"name" json:"name"`
		Namespace *string       `yaml:"namespace" json:"namespace"`
		Config    TaskConfigure `yaml:"config" json:"Config"`
		Depends   []Depend      `yaml:"depends" json:"depends"`
	}

	Depend struct {
		ID        string `yaml:"task" json:"task"`
		Condition *struct {
			Key      string            `yaml:"key" json:"key"`
			Value    interface{}       `yaml:"value" json:"value"`
			Operator ConditionOperator `yaml:"operator" json:"operator"`
		} `yaml:"condition" json:"condition"`
		Properties map[string]interface{} `yaml:"properties"`
	}
)
