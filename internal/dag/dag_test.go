package dag

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testVertex struct {
	id int
}

func (t testVertex) ID() string {
	return strconv.Itoa(t.id)
}

func createTestGraph() *acyclicGraph {
	g := &acyclicGraph{}
	a := g.Add(testVertex{1})
	b := g.Add(testVertex{2})
	c := g.Add(testVertex{3})
	g.Connect(BasicEdge(a, b))
	g.Connect(BasicEdge(a, c))
	return g
}

func TestLeaves(t *testing.T) {
	g := createTestGraph()
	leaves, err := g.Leaves()
	assert.NoError(t, err)
	assert.Len(t, leaves, 2)
	assert.ElementsMatch(t, leaves, []testVertex{{2}, {3}})
}

func TestRoot(t *testing.T) {
	g := createTestGraph()
	root, err := g.Root()
	assert.NoError(t, err)
	assert.NotNil(t, root)
	assert.Equal(t, root, testVertex{1})
}

func TestCycle(t *testing.T) {
	g := createTestGraph()
	assert.NoError(t, g.Validate())

	g.Connect(BasicEdge(testVertex{2}, testVertex{1}))
	assert.Error(t, g.Validate(), "no roots found")
}
