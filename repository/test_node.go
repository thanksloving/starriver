package repository

import (
	"errors"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/helper"
	"github.com/thanksloving/starriver/registry"
)

type (
	testParam struct {
		Pass bool `json:"pass" yaml:"pass"`
	}

	testNode struct {
		helper.SkeletonWithParameter
	}
)

func registerTestNode() {
	registry.Register("TestNode", "测试节点，可以明确指明是成功还是失败", func(id string) starriver.Executable {
		return &testNode{helper.NewSkeletonWithParameter(id, &testParam{})}
	})
}

func (t *testNode) Execute(dataContext starriver.DataContext, param interface{}) starriver.Response {
	p := param.(*testParam)
	dataContext.Infof("test node: %q, pass=%t", t.ID(), p.Pass)
	if p.Pass {
		return helper.NewSuccessResponse()
	} else {
		return helper.NewErrorResponse(errors.New("just test error"))
	}
}
