package repository

import (
	"reflect"
	"regexp"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/helper"
	"github.com/thanksloving/starriver/registry"
)

type (
	regexpComponent struct {
		helper.SkeletonWithParameter
	}

	RegexpParam struct {
		Expr    string
		Content string
	}
)

var _ starriver.Executable = (*regexpComponent)(nil)

func registerRegexp() {
	registry.Register("RegExp", "规则表达式执行",
		func(id string) starriver.Executable {
			return &regexpComponent{helper.NewSkeletonWithParameter(id, &RegexpParam{})}
		},
		registry.Input([]starriver.InputParam{
			{
				Key:      "Expr",
				Required: true,
				Desc:     "正则表达式",
			},
			{
				Key:      "Content",
				Required: true,
				Desc:     "待匹配的字符串",
			},
		}),
		registry.Output(map[string]starriver.OutputValue{
			"Result": {
				Desc: "匹配结果",
				Type: reflect.Slice,
			},
			"Match": {
				Desc: "是否命中",
				Type: reflect.Bool,
			},
		}),
	)
}

func (r *regexpComponent) ParameterNew() interface{} {
	return &RegexpParam{}
}

func (r *regexpComponent) Execute(_ starriver.DataContext, params interface{}) starriver.Response {
	param := params.(*RegexpParam)
	expr, err := regexp.Compile(param.Expr)
	if err != nil {
		return helper.NewErrorResponse(err)
	}
	result := expr.FindAllString(param.Content, -1)
	return helper.NewSuccessDataResponse(map[string]interface{}{
		"Result": result,
		"Match":  len(result) > 0,
	})
}
