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

	doc.GenManTree(cmd.RootCmd, header, ".")
	doc.GenMarkdownTree(cmd.RootCmd, ".")
}
