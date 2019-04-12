package main

import (
	"fmt"

	"github.com/wedeploy/cli/update/releasenotes"
)

var header = `---
title: "CLI"
description: "Check out the latest releases of the Liferay CLI Tool"
layout: "updates"
updates:
`

var entryTemplate = ` -
  version: %v
  date: %v
  description: %v
`

func main() {
	fmt.Print(header)

	for nc := len(releasenotes.ReleaseNotes) - 1; nc >= 0; nc-- {
		note := releasenotes.ReleaseNotes[nc]
		fmt.Printf(entryTemplate, note.Version, note.Date, note.Description)
	}

	fmt.Println("---")
}
