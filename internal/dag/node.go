package dag

const (
	NodeTypeAny NodeType = "any"
	// NodeTypeNot NotNode is a special node, it will pass when dependency node failed, so it can only have one dependency node
	NodeTypeNot NodeType = "not"
)

type (
	Node interface {
		GraphObject

		GetType() NodeType
	}

	NodeType string

	node struct {
		id    string
		types NodeType
	}
)

func NewNode(id string, nodeType NodeType) Node {
	return &node{
		id:    id,
		types: nodeType,
	}
}

func (n *node) GetType() NodeType {
	return n.types
}

func (n *node) ID() string {
	return n.id
}
