package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"

	"github.com/wedeploy/cli/context"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/user"
	"github.com/wedeploy/cli/verbose"
)

type Remote struct {
	URL     string
	Comment string
}

// Remotes (list of alternative endpoints)
type Remotes map[string]Remote

func (r Remotes) Set(name string, url string, comment ...string) {
	// make sure to use # by default, instead of ;
	if len(comment) != 0 {
		comment = append([]string{"#"}, comment...)
	}

	r[name] = Remote{
		URL:     url,
		Comment: strings.Join(comment, " "),
	}
}

func (r Remotes) Del(name string) {
	delete(r, name)
}

// Config of the application
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
	Remotes         Remotes   `ini:"-"`
	file            *ini.File `ini:"-"`
}

var (
	// Global configuration
	Global *Config

	// Context stores the environmental context
	Context *context.Context
)

// Load the configuration
func (c *Config) Load() {
	switch c.configExists() {
	case true:
		c.read()
	default:
		verbose.Debug("Config file not found.")
		c.file = ini.Empty()
		c.banner()
	}

	c.load()
}

// Save the configuration
func (c *Config) Save() {
	var cfg = c.file
	var err = cfg.ReflectFrom(c)

	if err != nil {
		panic(err)
	}

	c.updateRemotes()
	c.simplify()

	err = cfg.SaveTo(c.Path)

	if err != nil {
		panic(err)
	}
}

// Setup the environment
func Setup() {
	setupContext()
	setupGlobal()
}

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

func (c *Config) load() {
	c.setDefaults()

	if err := c.file.MapTo(c); err != nil {
		panic(err)
	}
}

func (c *Config) read() {
	var err error
	c.file, err = ini.Load(c.Path)

	if err != nil {
		println("Error reading configuration file:", err.Error())
		println("Fix " + c.Path + " by hand or erase it.")
		os.Exit(1)
	}

	c.readRemotes()
}

func (c *Config) readRemotes() {
	var remotes = c.file.Section("remotes")
	c.Remotes = Remotes{}

	for _, k := range remotes.KeyStrings() {
		key := remotes.Key(k)
		comment := strings.TrimPrefix(key.Comment, "#")

		c.Remotes[k] = Remote{
			URL:     key.Value(),
			Comment: strings.TrimSpace(comment),
		}
	}
}

func (c *Config) simplify() {
	var mainSection = c.file.Section("")
	var omitempty = []string{"next_version", "last_update_check"}

	for _, k := range omitempty {
		var key = mainSection.Key(k)
		if key.Value() == "" && key.Comment == "" {
			mainSection.DeleteKey(k)
		}
	}
}

func (c *Config) simplifyRemotes() {
	var remotes = c.file.Section("remotes")

	for _, k := range remotes.KeyStrings() {
		key := remotes.Key(k)

		if key.Value() == "" && key.Comment == "" {
			remotes.DeleteKey(k)
		}
	}

	if len(remotes.KeyStrings()) == 0 && remotes.Comment == "" {
		c.file.DeleteSection("remotes")
	}
}

func (c *Config) updateRemotes() {
	var remotes = c.file.Section("remotes")

	for _, k := range remotes.KeyStrings() {
		if _, ok := c.Remotes[k]; !ok {
			remotes.DeleteKey(k)
		}
	}

	for k, v := range c.Remotes {
		key := remotes.Key(k)
		key.SetValue(v.URL)
		key.Comment = v.Comment
	}

	c.simplifyRemotes()
}

func (c *Config) banner() {
	c.file.Section("DEFAULT").Comment = `# Configuration file for WeDeploy CLI
# https://wedeploy.io`
}

func setupContext() {
	var err error
	Context, err = context.Get()

	if err != nil {
		println(err.Error())
		os.Exit(-1)
	}
}

func setupGlobal() {
	Global = &Config{
		Path: filepath.Join(user.GetHomeDir(), ".we"),
	}

	Global.Load()
}

// Teardown resets the configuration environment
func Teardown() {
	teardownContext()
	teardownGlobal()
}

func teardownContext() {
	Context = nil
}

func teardownGlobal() {
	Global = nil
}
