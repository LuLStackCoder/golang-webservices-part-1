package tree

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

const (
	sep = string(os.PathSeparator)
	eof = "\n"

	defPrefix  = "│\t"
	lastPrefix = "\t"

	startLine = "├───"
	endLine   = "└───"
)

// Tree interface implemented the tree
type Tree interface {
	RenderTree(out io.Writer, prefix []string, nd *node)
	Root() *node
}

type tree struct {
	root *node
}

// Root returns the root
func (t *tree) Root() *node {
	return t.root
}

// RenderTree walk from whole tree and render the tree path recursively
func (t *tree) RenderTree(out io.Writer, prefix []string, node *node) {
	var lvlPrefix string
	for i, v := range node.children {
		if i == len(node.children)-1 {
			lvlPrefix = lastPrefix
			io.WriteString(out, strings.Join(prefix, ""))
			io.WriteString(out, endLine)
		} else {
			lvlPrefix = defPrefix
			io.WriteString(out, strings.Join(prefix, ""))
			io.WriteString(out, startLine)
		}
		io.WriteString(out, v.name)
		io.WriteString(out, eof)
		if v.isDir {
			t.RenderTree(out, append(prefix, lvlPrefix), v)
		}
	}
	prefix = prefix[:len(prefix)-1]
}

// NewTree is the constructor of tree
func NewTree(path string, withFiles bool) Tree {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil
	}
	root := newNode(path, true, buildTree(path, files, withFiles))
	return &tree{
		root: root,
	}
}

// buildTree build the path tree recursively
func buildTree(path string, files []os.FileInfo, withFiles bool) (nodes []*node) {
	for _, v := range files {
		var childFiles []os.FileInfo
		name := v.Name()
		isDir := v.IsDir()
		node := &node{}
		if isDir {
			childFiles, _ = ioutil.ReadDir(path + sep + name)
			node = newNode(name, isDir, buildTree(path+sep+name, childFiles, withFiles))
			nodes = append(nodes, node)
		} else if withFiles {
			size := v.Size()
			strSize := ""
			if size == 0 {
				strSize = " (empty)"
			} else {
				strSize = fmt.Sprintf(" (%db)", size)
			}
			node = newNode(name+strSize, isDir, nil)
			nodes = append(nodes, node)
		}
	}
	return
}
