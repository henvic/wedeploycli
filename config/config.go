package config

import (
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"

	"github.com/launchpad-project/cli/configstore"
	"github.com/launchpad-project/cli/context"
	"github.com/launchpad-project/cli/defaults"
	"github.com/launchpad-project/cli/user"
	"github.com/launchpad-project/cli/verbose"
)

type Config struct {
	Username        string    `ini:"username"`
	Password        string    `ini:"password"`
	Token           string    `ini:"token"`
	Local           bool      `ini:"local"`
	NoColor         bool      `ini:"disable_colors"`
	Endpoint        string    `ini:"endpoint"`
	NotifyUpdates   bool      `ini:"notify_updates"`
	ReleaseChannel  string    `ini:"release_channel"`
	LastUpdateCheck string    `ini:"last_update_check"`
	NextVersion     string    `ini:"next_version"`
	Path            string    `ini:"-"`
	file            *ini.File `ini:"-"`
}

var UserConfigurableKeys = []string{
	"username",
	"password",
	"endpoint",
}

var (
	Global *Config

	// Context stores the environmental context
	Context *context.Context

	// Stores sets the map of available config stores
	Stores = map[string]*configstore.Store{}
)

func (c *Config) setDefaults() {
	c.Local = true
	c.Endpoint = defaults.Endpoint
	c.NotifyUpdates = true
	c.ReleaseChannel = "stable"
}

func (c *Config) configExists() bool {
	var _, err = os.Stat(c.Path)

	switch {
	case err == nil:
		return true
	case os.IsNotExist(err):
		return false
	default:
		panic(err)
	}
}

func (c *Config) read() {
	var cfg, err = ini.LooseLoad(c.Path)
	c.file = cfg

	if !c.configExists() {
		verbose.Debug("Config file not found.")
		c.banner()
		return
	}

	if err != nil {
		println("Error reading configuration file:", err.Error())
		println("Fix " + c.Path + " by hand or erase it.")
		os.Exit(1)
	}

}

func (c *Config) banner() {
	c.file.Section("DEFAULT").Comment = `# Configuration file for WeDeploy CLI
# https://wedeploy.io`
}

func (c *Config) Load() {
	c.setDefaults()
	c.read()

	if err := c.file.MapTo(c); err != nil {
		panic(err)
	}
}

func (c *Config) Save() {
	var err = c.file.ReflectFrom(c)

	if err != nil {
		panic(err)
	}

	if c.NextVersion == "" {
		c.file.Section("").DeleteKey("next_version")
	}

	err = c.file.SaveTo(c.Path)

	if err != nil {
		panic(err)
	}
}

// Setup the environment
func Setup() {
	Stores = map[string]*configstore.Store{}

	Global = &Config{
		Path: filepath.Join(user.GetHomeDir(), "/.we"),
	}

	Global.Load()

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
	Global = nil
	Stores = map[string]*configstore.Store{}
}
