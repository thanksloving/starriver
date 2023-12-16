package dag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraph(t *testing.T) {
	g := createTestGraph()
	assert.Equal(t, true, g.HasEdge(BasicEdge(testVertex{1}, testVertex{2})))
	assert.Equal(t, false, g.HasEdge(BasicEdge(testVertex{2}, testVertex{3})))

	assert.Equal(t, true, g.HasVertex(testVertex{1}))
	assert.Equal(t, false, g.HasVertex(testVertex{100}))

	edges := g.EdgesFrom(testVertex{1})
	assert.ElementsMatch(t, edges, []Edge{
		BasicEdge(testVertex{1}, testVertex{2}),
		BasicEdge(testVertex{1}, testVertex{3}),
	})

	edges = g.EdgesTo(testVertex{2})
	assert.ElementsMatch(t, edges, []Edge{
		BasicEdge(testVertex{1}, testVertex{2}),
	})

}
