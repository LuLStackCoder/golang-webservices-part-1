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

	defPrefix  = "│  "
	lastPrefix = "  "

	startLine = "├───"
	endLine   = "└───"
)

// Tree
type Tree interface {
	RenderTree(out io.Writer, prefix []string, nd *node)
	Root() *node
}

type tree struct {
	root *node
}

// Root ...
func (t *tree) Root() *node {
	return t.root
}

// RenderTree ...
func (t *tree) RenderTree(out io.Writer, prefix []string, nd *node) {
	var lvlPrefix string
	for i, v := range nd.children {
		if i == len(nd.children)-1 {
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
	root := newNode(path, true, BuildTree(path, files, withFiles))
	return &tree{
		root: root,
	}
}

// BuildTree ...
func BuildTree(path string, files []os.FileInfo, withFiles bool) (nodes []*node) {
	for _, v := range files {
		name := v.Name()
		isDir := v.IsDir()
		var childFiles []os.FileInfo
		node := &node{}
		if isDir {
			childFiles, _ = ioutil.ReadDir(path + sep + name)
			node = newNode(name, isDir, BuildTree(path+sep+name, childFiles, withFiles))
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
