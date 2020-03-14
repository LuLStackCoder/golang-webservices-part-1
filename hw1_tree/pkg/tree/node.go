package tree

type node struct {
	name     string
	isDir    bool
	children []*node
}

// newNode ...
func newNode(name string, isDir bool, children []*node) *node {
	return &node{
		name:     name,
		isDir:    isDir,
		children: children,
	}
}
