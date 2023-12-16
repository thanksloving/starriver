package repository

import (
	"bytes"
	"text/template"
	"time"

	"github.com/spf13/cast"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/helper"
	"github.com/thanksloving/starriver/registry"
)

type (
	templateComponent struct {
		helper.SkeletonWithParameter
	}

	templateParam struct {
		Template  string
		OutputKey string
		Shared    bool
	}
)

var _ starriver.Executable = (*templateComponent)(nil)

func registerTemplate() {
	registry.Register("Template", "文本模板填充",
		func(id string) starriver.Executable {
			return &templateComponent{helper.NewSkeletonWithParameter(id, &templateParam{})}
		},
		registry.Input([]starriver.InputParam{
			{
				Key:      "Template",
				Required: true,
				Desc:     "模板，变量使用 {{ str \\\"test\\\" }} 占位，必须使用 {{ str }} 的形式",
			},
			{
				Key:      "OutputKey",
				Required: true,
				Desc:     "组件输出的结果名称，下一个节点可以通过这个 key 获取到值\" required:\"true\"",
			},
			{
				Key:      "Shared",
				Required: true,
				Desc:     "是否共享该结果，默认不共享，只有依赖的节点能拿到值",
			},
		}),
		registry.Output(map[string]starriver.OutputValue{}),
	)
}

func (tc *templateComponent) Execute(dataContext starriver.DataContext, param interface{}) starriver.Response {
	p := param.(*templateParam)
	t := template.New(tc.ID()).Funcs(map[string]any{
		"str": func(key string) string {
			val, _ := dataContext.Get(key)
			return cast.ToString(val)
		},
		"now": func() string {
			return time.Now().Format("2006-01-02 15:04:05")
		},
	})
	buf := new(bytes.Buffer)
	if err := template.Must(t.Parse(p.Template)).Execute(buf, struct{}{}); err != nil {
		return helper.NewErrorResponse(err)
	}
	content := buf.String()
	if p.Shared {
		dataContext.Set(p.OutputKey, content)
	}
	return helper.NewSuccessDataResponse(map[string]interface{}{
		p.OutputKey: content,
	})
}
