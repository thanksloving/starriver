package repository

import (
	"github.com/thanksloving/starriver"
	"github.com/thanksloving/starriver/helper"
	"github.com/thanksloving/starriver/registry"
)

type (
	ReShareDataNode struct {
		helper.SkeletonWithParameter
	}

	restoreData struct {
		Data map[string]interface{}
	}
)

func registerReShareDataNode() {
	registry.Register("ReShareDataNode", "将指定的数据转存至共享工作区，数据可能是前置节点的输出，边的属性，也可能是已经存在于工作区的数据（更换 key 的场景）",
		func(id string) starriver.Executable {
			return &ReShareDataNode{helper.NewSkeletonWithParameter(id, restoreData{})}
		},
		registry.Input([]starriver.InputParam{
			{
				Key:      "Data",
				Required: true,
				Desc:     "需要转存的 key 以及工作区的新 key",
			},
		}),
	)
}

func (spd *ReShareDataNode) Execute(dc starriver.DataContext, params interface{}) starriver.Response {
	data := params.(*restoreData)
	for key, sharedKey := range data.Data {
		value, ok := dc.Get(key)
		if !ok {
			dc.Warnf("ShareData not exist, key=%q", key)
			continue
		}
		newKey, ok := sharedKey.(string)
		if !ok {
			dc.Warnf("ShareData key not string, key=%v, type=%T", newKey, newKey)
			continue
		}
		if v, ok := dc.Get(sharedKey.(string)); ok {
			dc.Warnf("ShareData with key %q will be override from %v to %v", sharedKey, v, value)
		}
		ok = dc.Set(sharedKey.(string), value)
		dc.Debugf("ReShare key=%q to key=%q %v, value=%v", key, newKey, ok, value)
	}
	return helper.NewSuccessResponse()
}
