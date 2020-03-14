package main

import (
	"io"
	"os"

	"github.com/LuLStackCoder/golang-webservices-part-1/hw1_tree/pkg/tree"
)

// dirTree is a function that create tree and render in in output
func dirTree(out io.Writer, path string, printFiles bool) (err error) {
	fileTree := tree.NewTree(path, printFiles)
	fileTree.RenderTree(out, []string{""}, fileTree.Root())
	return
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
