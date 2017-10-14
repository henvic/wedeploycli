package main

import (
	"github.com/spf13/cobra/doc"
	"github.com/wedeploy/cli/cmd/root"
)

func main() {
	header := &doc.GenManHeader{
		Title:  "WeDeploy CLI Tool",
		Source: "Docs created automatically from the source files",
	}

	if err := doc.GenManTree(root.Cmd, header, "."); err != nil {
		panic(err)
	}

	if err := doc.GenMarkdownTree(root.Cmd, "."); err != nil {
		panic(err)
	}
}
