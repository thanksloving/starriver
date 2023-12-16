package dag

import (
	"fmt"
	"reflect"

	"github.com/spf13/cast"

	"github.com/thanksloving/starriver"
)

// Edge represents an edge in the graph, with a source and target vertex.
type (
	Edge interface {
		GraphObject

		Source() Vertex
		Target() Vertex
		Properties() map[string]interface{}
		WithProperties(map[string]interface{}) Edge
	}

	IsConditionalEdge interface {
		Match(ctx starriver.DataContext) bool
	}

	// basicEdge is a basic implementation of Edge that has the source and
	// target vertex.
	basicEdge struct {
		S, T       Vertex
		properties map[string]interface{}
	}

	conditionEdge struct {
		basicEdge
		key      string
		value    interface{}
		operator starriver.ConditionOperator
	}
)

// BasicEdge returns an Edge implementation that simply tracks the source
// and target given as-is.
func BasicEdge(source, target Vertex) Edge {
	return &basicEdge{S: source, T: target}
}

// ConditionEdge return an Edge with Condition
func ConditionEdge(source, target Vertex, key string, value interface{}, operator starriver.ConditionOperator) Edge {
	return &conditionEdge{
		basicEdge: basicEdge{
			S: source, T: target,
		},
		key:      key,
		value:    value,
		operator: operator,
	}
}

// PropertyEdge return an Edge with Property
func PropertyEdge(source, target Vertex, properties map[string]interface{}) Edge {
	return &basicEdge{
		S: source, T: target,
		properties: properties,
	}
}

func (e *basicEdge) Source() Vertex {
	return e.S
}

func (e *basicEdge) Target() Vertex {
	return e.T
}

func (e *basicEdge) ID() string {
	return e.S.ID() + "->" + e.T.ID()
}

func (e *basicEdge) WithProperties(properties map[string]interface{}) Edge {
	if e.properties == nil {
		e.properties = properties
		return e
	}
	for k, v := range properties {
		e.properties[k] = v
	}
	return e
}

func (c *conditionEdge) Match(dc starriver.DataContext) bool {
	val, ok := dc.Get(c.key)
	if !ok {
		dc.Errorf("condition eval fail, key=%v not exist", c.key)
		return false
	}
	switch c.operator {
	case starriver.ConditionEQ:
		return reflect.DeepEqual(val, c.value)
	case starriver.ConditionNE:
		return !reflect.DeepEqual(val, c.value)
	case starriver.ConditionIn:
		s := reflect.ValueOf(c.value)
		if s.Kind() != reflect.Slice {
			return false
		}
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) {
				return true
			}
		}
		return false
	case starriver.ConditionGT, starriver.ConditionLT, starriver.ConditionGE, starriver.ConditionLE:
		result, err := compareNumberValue(val, c.value, c.operator)
		if err != nil {
			dc.Errorf("condition eval error, source=%v, target=%v, cause:%v", val, c.value, err)
			return false
		}
		return result
	}
	return false
}

func compareNumberValue(source, target interface{}, operator starriver.ConditionOperator) (bool, error) {
	s, e := cast.ToFloat64E(source)
	if e != nil {
		return false, e
	}
	t, e := cast.ToFloat64E(target)
	if e != nil {
		return false, e
	}
	switch operator {
	case starriver.ConditionGT:
		return s > t, nil
	case starriver.ConditionLT:
		return s < t, nil
	case starriver.ConditionGE:
		return s >= t, nil
	case starriver.ConditionLE:
		return s <= t, nil
	default:
		return false, fmt.Errorf("%q not support", operator)
	}
}

func (e *basicEdge) Properties() map[string]interface{} {
	return e.properties
}
