package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/thanksloving/starriver"
)

type dummyParam struct {
	Name string
	Age  int
}

func TestPrepareParameter_NilValue(t *testing.T) {
	p := &mockPipeline{}
	dc := NewDataContext(context.Background(), p, nil)
	ap := &assembleParam{id: "test"}

	// When a variable is not found and required is false, it returns nil.
	// Without the fix, this would panic.
	params := starriver.Params{
		{
			Name:     "Name",
			Type:     starriver.ParamTypeVariable,
			Variable: "not_exist",
			Required: false,
		},
		{
			Name:    "Age",
			Type:    starriver.ParamTypeLiteral,
			Literal: 18,
		},
	}

	obj := &dummyParam{}
	res, err := ap.prepareParameter(dc, params, obj)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	
	resObj := res.(*dummyParam)
	assert.Equal(t, "", resObj.Name)
	assert.Equal(t, 18, resObj.Age)
}
