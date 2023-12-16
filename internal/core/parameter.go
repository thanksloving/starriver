package core

import (
	"fmt"
	"reflect"

	dafaults "github.com/mcuadros/go-defaults"

	"github.com/thanksloving/starriver"
)

type assembleParam struct {
	id string
}

func (ap *assembleParam) prepareParameter(dataContext starriver.DataContext, paramConfigs starriver.Params, paramObj interface{}) (param interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("prepare parameter error, %v", r)
		}
	}()
	dafaults.SetDefaults(paramObj)
	v := reflect.ValueOf(paramObj).Elem()
	if !v.CanAddr() {
		return nil, fmt.Errorf("cannot assign to the item passed, item must be a pointer in order to assign")
	}
	for _, paramConfig := range paramConfigs {
		val, err := ap.getValue(dataContext, paramConfig)
		if err != nil {
			return nil, err
		}
		if field := v.FieldByName(paramConfig.Name); field.CanAddr() {
			field.Set(reflect.ValueOf(val))
		} else {
			err = fmt.Errorf("[PrepareParameter]id= %q parameter %q init failed, config=%+v", ap.id, paramConfig.Name, paramConfig)
			return nil, err
		}
	}
	return paramObj, nil
}

func (ap *assembleParam) getValue(dataContext starriver.DataContext, paramConfig starriver.Param) (val interface{}, err error) {
	switch paramConfig.Type {
	case starriver.ParamTypeVariable:
		var ok bool
		if val, ok = dataContext.Get(paramConfig.Variable); !ok && paramConfig.Required {
			err = fmt.Errorf("[PrepareParameter]id=%q get %q failed", ap.id, paramConfig.Variable)
		}
	case starriver.ParamTypeLiteral:
		val = paramConfig.Literal
	case starriver.ParamTypeComplex:
		val = make([]interface{}, len(paramConfig.Complex))
		for idx, item := range paramConfig.Complex {
			if v, e := ap.getValue(dataContext, item); e != nil {
				return nil, e
			} else {
				val.([]interface{})[idx] = v
			}
		}
	case starriver.ParamTypeMapping:
		val = make(map[string]interface{})
		for key, param := range paramConfig.Mapping {
			if v, e := ap.getValue(dataContext, param); e != nil {
				return nil, e
			} else {
				val.(map[string]interface{})[key] = v
			}
		}
	default:
		if paramConfig.Required {
			err = fmt.Errorf("[PrepareParameter]id=%q param type is not support, %v", ap.id, paramConfig)
		}
	}
	return val, err
}
