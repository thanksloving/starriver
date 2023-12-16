package repository

import (
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"github.com/spf13/cast"

	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/helper"
	"github.com/thanksloving/starriver/registry"
)

type (
	JsonPath struct {
		helper.SkeletonWithParameter
	}

	jsonParam struct {
		JsonString string
		Exprs      map[string]interface{}
	}
)

var _ starriver.Executable = (*JsonPath)(nil)

func registerJsonPath() {

	registry.Register("JsonPath", "Json 抽取，采用 JsonPath 语法", func(id string) starriver.Executable {
		return &JsonPath{helper.NewSkeletonWithParameter(id, &jsonParam{})}
	},
		registry.Input([]starriver.InputParam{
			{
				Key:      "JsonString",
				Desc:     "json 字符串",
				Required: true,
			},
			{
				Key:  "Exprs",
				Desc: "Map, key 是输出结果名，Value 是解析抽取的表达式",
			},
		}),
	)
}

func (j *JsonPath) Execute(dataContext starriver.DataContext, param interface{}) starriver.Response {
	p := param.(*jsonParam)
	obj, err := oj.ParseString(p.JsonString)
	if err != nil {
		dataContext.Errorf("[%q]not a valid json, err=%v, json=%q", j.ID(), err, p.JsonString)
		return helper.NewErrorResponse(err)
	}
	result := make(map[string]interface{})
	for k, v := range p.Exprs {
		x, e := jp.ParseString(cast.ToString(v))
		if e != nil {
			dataContext.Errorf("[%q]expr not valid, err=%v, expr=%q", j.ID(), e, v)
			return helper.NewErrorResponse(e)
		}
		result[k] = x.Get(obj)
	}
	return helper.NewSuccessDataResponse(result)
}
