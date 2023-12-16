package helper

import (
	"reflect"

	"github.com/thanksloving/starriver"
)

type (
	Skeleton interface {
		starriver.Executable
	}

	basicSkeleton struct {
		Id string
		Skeleton
	}

	SkeletonWithParameter interface {
		starriver.Executable
		starriver.WithParameters
	}

	identityWithParameters struct {
		basicSkeleton
		value reflect.Value
		starriver.WithParameters
	}
)

func (i *basicSkeleton) ID() string {
	return i.Id
}

func NewSkeleton(id string) Skeleton {
	return &basicSkeleton{Id: id}
}

func NewSkeletonWithParameter(id string, param interface{}) SkeletonWithParameter {
	paramType := reflect.TypeOf(param)
	var value reflect.Value
	switch paramType.Kind() {
	case reflect.Pointer:
		elem := paramType.Elem()
		value = reflect.New(elem)
	case reflect.Struct:
		value = reflect.New(paramType)
	default:
		panic("param must be struct or pointer to struct")
	}
	return &identityWithParameters{
		basicSkeleton: basicSkeleton{Id: id},
		value:         value,
	}
}

func (iwp *identityWithParameters) ParameterNew() interface{} {
	return iwp.value.Interface()
}
