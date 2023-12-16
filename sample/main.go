package main

import (
	"context"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/thanksloving/starriver/flow"
)

func main() {
	pwd, _ := os.Getwd()
	config, err := os.ReadFile(filepath.Join(pwd, "/sample/demo.yml"))
	if err != nil {
		panic(err)
	}
	pc, err := flow.LoadPipelineByYaml(string(config))
	if err != nil {
		panic(err)
	}
	sharedDataStore := flow.NewSharedDataStore()
	pipeline, err := flow.NewPipeline(*pc)
	if err != nil {
		panic(err)
	}
	dc := flow.NewDataContext(context.Background(), pipeline, map[string]interface{}{
		"question": "今天天气怎么样？",
		"answer":   "你瞎吗，为什么不自己看？",
	}, flow.SetSharedDataStore(sharedDataStore))
	engine := flow.NewRiverEngine()
	defer engine.Destroy()
	result := engine.Run(dc, pipeline)
	log.Infof("Data=%+v, Status=%s, State=%+v", result.Data, result.Status, result.State)
}
