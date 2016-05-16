package main

import (
	"github.com/launchpad-project/cli/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	header := &doc.GenManHeader{
		Title:  "WeDeploy CLI Tool",
		Source: "Docs created automatically from the source files",
	}

	doc.GenManTree(cmd.RootCmd, header, ".")
	doc.GenMarkdownTree(cmd.RootCmd, ".")
}
