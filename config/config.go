package config

import (
	"os"
	"path/filepath"

	"github.com/launchpad-project/cli/configstore"
	"github.com/launchpad-project/cli/context"
	"github.com/launchpad-project/cli/user"
)

var (
	// Context stores the environmental context
	Context *context.Context

	// Stores sets the map of available config stores
	Stores = map[string]*configstore.Store{}
)

// Setup the environment
func Setup() {
	Stores = map[string]*configstore.Store{}

	Stores["global"] = &configstore.Store{
		Name: "global",
		Path: user.GetHomeDir() + "/.launchpad.json",
		ConfigurableKeys: []string{
			"username",
			"password",
			"endpoint",
		},
	}

	var err error
	Context, err = context.Get()

	if err != nil {
		println(err.Error())
		os.Exit(-1)
	}

	if Context.Scope == "project" || Context.Scope == "container" {
		Stores["project"] = &configstore.Store{
			Name: "project",
			ConfigurableKeys: []string{
				"id",
				"name",
				"description",
				"domain",
			},
			Path: filepath.Join(Context.ProjectRoot, "/project.json"),
		}
	}

	if Context.Scope == "container" {
		Stores["container"] = &configstore.Store{
			Name: "container",
			Path: filepath.Join(Context.ContainerRoot, "/container.json"),
		}
	}

	for k := range Stores {
		err := Stores[k].Load()

		if err != nil && !os.IsNotExist(err) {
			println("Unexpected error reading configuration file.")
			println("Fix " + Stores[k].Path + " by hand or erase it.")
			os.Exit(1)
		}
	}
}

// Teardown resets the configuration environment
func Teardown() {
	Context = nil
	Stores = map[string]*configstore.Store{}
}
