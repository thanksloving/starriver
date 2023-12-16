package dag

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/thanksloving/starriver"
)

type testDataContext struct {
	data map[string]interface{}
}

func (t testDataContext) Env(s string) (interface{}, bool) {
	panic("implement me")
}

func (t testDataContext) WithTimeout(timeout time.Duration) context.CancelFunc {
	_, cancel := context.WithTimeout(t, timeout)
	return cancel
}

func (t testDataContext) Stop() {}

func (t testDataContext) Context() context.Context {
	panic("implement me")
}

func (t testDataContext) GetRequestID() string {
	panic("implement me")
}

func (t testDataContext) Pipeline() starriver.Pipeline {
	panic("implement me")
}

func (t testDataContext) Release() {
	panic("implement me")
}

func (t testDataContext) Debug(msg string) {
	panic("implement me")
}

func (t testDataContext) Debugf(msg string, args ...interface{}) {
	panic("implement me")
}

func (t testDataContext) Info(msg string) {
	panic("implement me")
}

func (t testDataContext) Infof(msg string, args ...interface{}) {
	panic("implement me")
}

func (t testDataContext) Warn(msg string) {
	panic("implement me")
}

func (t testDataContext) Warnf(msg string, args ...interface{}) {
	panic("implement me")
}

func (t testDataContext) Error(msg string) {
	panic("implement me")
}

func (t testDataContext) Errorf(msg string, args ...interface{}) {

}

func (t testDataContext) Fatal(msg string) {
	panic("implement me")
}

func (t testDataContext) Fatalf(msg string, args ...interface{}) {
	panic("implement me")
}

func (t testDataContext) Set(key string, val interface{}) bool {
	panic("implement me")
}

func (t testDataContext) Del(key string) {
	panic("implement me")
}

func (t testDataContext) Get(key string) (interface{}, bool) {
	v, ok := t.data[key]
	return v, ok
}

func (t testDataContext) Marshal() ([]byte, error) {
	panic("implement me")
}

func (t testDataContext) Configure(flowName, requestID string) {
	panic("implement me")
}

func (t testDataContext) SetCurrentNodeData(nodeId string, data map[string]interface{}) {
	panic("implement me")
}

func (t testDataContext) GetDependNodeValue(nodeId, key string) (interface{}, bool) {
	panic("implement me")
}

func (t testDataContext) Deadline() (deadline time.Time, ok bool) {
	panic("implement me")
}

func (t testDataContext) Done() <-chan struct{} {
	panic("implement me")
}

func (t testDataContext) Err() error {
	panic("implement me")
}

func (t testDataContext) Value(key any) any {
	panic("implement me")
}

func (t testDataContext) WithValue(key, value any) {
	panic("implement me")
}

func TestConditionEdge(t *testing.T) {
	g := &acyclicGraph{}
	a := g.Add(testVertex{1})
	b := g.Add(testVertex{2})
	dataContext := testDataContext{
		data: map[string]interface{}{
			"tag":   "C4",
			"age":   18,
			"price": 988.5,
		},
	}
	tests := []struct {
		key    string
		value  interface{}
		logic  starriver.ConditionOperator
		result bool
	}{
		{"k", "", "x", false},
		{"tag", "C3", "==", false},
		{"tag", "C4", "==", true},
		{"tag", "C4", "!=", false},
		{"tag", []string{"C4"}, "in", true},
		{"tag", []int{1}, "in", false},
		{"age", 18, ">", false},
		{"age", 18, ">=", true},
		{"age", 19.23, "<=", true},
		{"price", 800, "<=", false},
		{"price", 988.5, ">=", true},
		{"price", 988, ">", true},
		{"price", 988, "!=", true},
		{"tag", 988, "!=", true},
		{"tag", "C5", "!=", true},
		{"tag", "c5", "!=", true},
	}
	for _, test := range tests {
		edge := ConditionEdge(a, b, test.key, test.value, test.logic)
		result := edge.(*conditionEdge).Match(dataContext)
		assert.Equal(t, test.result, result)
	}
}
