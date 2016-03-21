package info

import (
	"fmt"
	"os"

	"github.com/launchpad-project/cli/config"
)

// Print prints information from project and containers
func Print() {
	var ctx = config.Context

	switch ctx.Scope {
	case "project":
		PrintProject()
	case "container":
		PrintContainer()
	default:
		println("fatal: not a project")
		os.Exit(1)
	}
}

// PrintProject prints project information read from project.json
func PrintProject() {
	var s = config.Stores["project"]

	fmt.Println(
		"Project: " + s.Get("id") + " (" + s.Get("name") + ")\n" +
			"Domain: " + s.Get("domain") + "\n" +
			"Description: " + s.Get("description"))
}

// PrintContainer prints container information read from container.json
func PrintContainer() {
	var s = config.Stores["container"]

	fmt.Println(
		"Container: " + s.Get("name") + "\n" +
			"Description: " + s.Get("description") + "\n" +
			"Version: " + s.Get("version") + "\n" +
			"Runtime: " + s.Get("runtime"))
}
