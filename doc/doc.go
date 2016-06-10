package main

import (
	"github.com/spf13/cobra/doc"
	"github.com/wedeploy/cli/cmd"
)

func main() {
	header := &doc.GenManHeader{
		Title:  "WeDeploy CLI Tool",
		Source: "Docs created automatically from the source files",
	}

	if err := doc.GenManTree(cmd.RootCmd, header, "."); err != nil {
		panic(err)
	}

	if err := doc.GenMarkdownTree(cmd.RootCmd, "."); err != nil {
		panic(err)
	}
}
