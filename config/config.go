package config

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/hashicorp/errwrap"
	version "github.com/hashicorp/go-version"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/remotes"
	"github.com/wedeploy/cli/verbose"
	"gopkg.in/ini.v1"
)

// Config of the application
type Config struct {
	Path string

	Params Params

	m    sync.Mutex
	file *ini.File
}

// GetParams gets the configuration parameters in a concurrent safe way.
func (c *Config) GetParams() Params {
	c.m.Lock()
	defer c.m.Unlock()
	return c.Params
}

// SetParams updates the set of params.
func (c *Config) SetParams(p Params) {
	c.m.Lock()
	defer c.m.Unlock()
	c.Params = p
}

// Params for the configuration
type Params struct {
	DefaultRemote   string        `ini:"default_remote"`
	NoAutocomplete  bool          `ini:"disable_autocomplete_autoinstall"`
	NoColor         bool          `ini:"disable_colors"`
	NotifyUpdates   bool          `ini:"notify_updates"`
	ReleaseChannel  string        `ini:"release_channel"`
	LastUpdateCheck string        `ini:"last_update_check"`
	PastVersion     string        `ini:"past_version"`
	NextVersion     string        `ini:"next_version"`
	EnableAnalytics bool          `ini:"enable_analytics"`
	AnalyticsID     string        `ini:"analytics_id"`
	EnableCURL      bool          `ini:"enable_curl"`
	Remotes         *remotes.List `ini:"-"`
}

var parseRemoteSectionNameRegex = regexp.MustCompile(`remote \"(.*)\"`)

// Load the configuration
func (c *Config) Load() error {
	if err := c.openOrCreate(); err != nil {
		return err
	}

	c.load()

	var params = c.GetParams()

	if params.Remotes == nil {
		params.Remotes = &remotes.List{}
	}

	c.SetParams(params)

	c.loadDefaultRemotes()
	return c.validateDefaultRemote()
}

func (c *Config) openOrCreate() error {
	if c.configExists() {
		return c.read()
	}

	verbose.Debug("Config file not found.")
	c.file = ini.Empty()
	c.banner()
	return nil
}

func (c *Config) loadDefaultRemotes() {
	var params = c.GetParams()
	var rl = params.Remotes

	switch rl.Get(defaults.CloudRemote).Infrastructure {
	case defaults.Infrastructure:
	case "":
		rl.Set(defaults.CloudRemote, remotes.Entry{
			Infrastructure:        defaults.Infrastructure,
			InfrastructureComment: "Default cloud remote",
		})
	default:
		_, _ = fmt.Fprintln(os.Stderr, color.Format(
			color.FgHiRed,
			"Warning: Non-standard wedeploy remote cloud detected"))
	}
}

func (c *Config) validateDefaultRemote() error {
	var params = c.Params
	var rl = params.Remotes
	var keys = rl.Keys()

	for _, k := range keys {
		if params.DefaultRemote == k {
			return nil
		}
	}

	return fmt.Errorf(`Remote "%v" is set as default, but not found.
Please fix your ~/.lcp file`, c.Params.DefaultRemote)
}

// Save the configuration
func (c *Config) Save() error {
	var cfg = c.file
	var params = c.GetParams()
	var err = cfg.ReflectFrom(&params)

	if err != nil {
		return errwrap.Wrapf("can't load configuration: {{err}}", err)
	}

	c.updateRemotes()
	c.simplify()

	err = cfg.SaveToIndent(c.Path, "    ")

	if err != nil {
		return errwrap.Wrapf("can't save configuration: {{err}}", err)
	}

	return nil
}

func (c *Config) setDefaults() {
	var params = c.GetParams()

	params.EnableAnalytics = true
	params.NotifyUpdates = true
	params.ReleaseChannel = defaults.StableReleaseChannel
	params.DefaultRemote = defaults.CloudRemote

	// By design Windows users should see no color unless they enable it
	// Issue #51.
	// https://github.com/wedeploy/cli/issues/51
	if runtime.GOOS == "windows" {
		params.NoColor = true
	}

	c.SetParams(params)
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

	var params = c.GetParams()

	if err := c.file.MapTo(&params); err != nil {
		panic(err)
	}

	c.SetParams(params)
}

func (c *Config) read() error {
	var err error
	c.file, err = ini.Load(c.Path)

	if err != nil {
		return errwrap.Wrapf("error reading configuration file: {{err}}\n"+
			"Fix "+c.Path+" by hand or erase it.", err)
	}

	c.readRemotes()
	c.checkNextVersionCacheIsNewer()

	return nil
}

func (c *Config) checkNextVersionCacheIsNewer() {
	if defaults.Version == "master" {
		return
	}

	vThis, err := version.NewVersion(defaults.Version)

	if err != nil {
		verbose.Debug(err)
		return
	}

	var params = c.GetParams()

	vNext, err := version.NewVersion(params.NextVersion)

	if err != nil {
		verbose.Debug(err)
		return
	}

	if vThis.GreaterThan(vNext) {
		params.NextVersion = ""
		c.SetParams(params)
	}
}

func parseRemoteSectionName(parsable string) (parsed string, is bool) {
	var matches = parseRemoteSectionNameRegex.FindStringSubmatch(parsable)

	if len(matches) == 2 {
		parsed = matches[1]
		is = true
	}

	return parsed, is
}

func (c *Config) listRemotes() []string {
	var list = []string{}

	for _, k := range c.file.SectionStrings() {
		var key, is = parseRemoteSectionName(k)

		if !is {
			continue
		}

		list = append(list, key)
	}

	return list
}

func (c *Config) getRemote(key string) *ini.Section {
	return c.file.Section(`remote "` + key + `"`)
}

func (c *Config) deleteRemote(key string) {
	c.file.DeleteSection(`remote "` + key + `"`)
}

func (c *Config) readRemotes() {
	var rl = &remotes.List{}

	for _, k := range c.listRemotes() {
		remote := c.getRemote(k)
		infrastructure := remote.Key("infrastructure")
		service := remote.Key("service")
		username := remote.Key("username")
		token := remote.Key("token")
		comment := remote.Comment

		rl.Set(k, remotes.Entry{
			Comment:               strings.TrimSpace(comment),
			Infrastructure:        strings.TrimSpace(infrastructure.String()),
			InfrastructureComment: strings.TrimSpace(infrastructure.Comment),
			Service:               strings.TrimSpace(service.String()),
			ServiceComment:        strings.TrimSpace(service.Comment),
			Username:              strings.TrimSpace(username.String()),
			UsernameComment:       strings.TrimSpace(username.Comment),
			Token:                 strings.TrimSpace(token.String()),
			TokenComment:          strings.TrimSpace(token.Comment),
		})
	}

	var params = c.GetParams()
	params.Remotes = rl
	c.SetParams(params)
}

func (c *Config) simplify() {
	var mainSection = c.file.Section("")
	var omitempty = []string{"past_version", "next_version", "last_update_check"}

	for _, k := range omitempty {
		var key = mainSection.Key(k)
		if key.Value() == "" && key.Comment == "" {
			mainSection.DeleteKey(k)
		}
	}
}

func (c *Config) simplifyRemotes() {
	for _, k := range c.listRemotes() {
		remote := c.getRemote(k)
		infrastructure := remote.Key("infrastructure")
		service := remote.Key("service")

		if infrastructure.Value() == "" && infrastructure.Comment == "" {
			remote.DeleteKey("infrastructure")
		}

		if service.Value() == "" && service.Comment == "" {
			remote.DeleteKey("service")
		}

		if len(remote.Keys()) == 0 && remote.Comment == "" {
			c.deleteRemote(k)
		}

		var omitempty = []string{"username", "token"}

		for _, k := range omitempty {
			var key = remote.Key(k)
			if key.Value() == "" && key.Comment == "" {
				remote.DeleteKey(k)
			}
		}
	}
}

func (c *Config) updateRemotes() {
	var params = c.GetParams()
	var rl = params.Remotes

	for _, k := range c.listRemotes() {
		if !rl.Has(k) {
			c.deleteRemote(k)
		}
	}

	for _, k := range rl.Keys() {
		var v = rl.Get(k)
		remote := c.getRemote(k)

		keyInfrastructure := remote.Key("infrastructure")
		keyInfrastructure.SetValue(v.Infrastructure)
		keyInfrastructure.Comment = v.InfrastructureComment

		keyService := remote.Key("service")
		keyService.SetValue(v.Service)
		keyService.Comment = v.ServiceComment

		keyUsername := remote.Key("username")
		keyUsername.SetValue(v.Username)
		keyUsername.Comment = v.UsernameComment

		keyToken := remote.Key("token")
		keyToken.SetValue(v.Token)
		keyToken.Comment = v.TokenComment

		remote.Comment = v.Comment
	}

	c.simplifyRemotes()
}

func (c *Config) banner() {
	c.file.Section("DEFAULT").Comment = `; Configuration file for Liferay Cloud
; https://www.liferay.com/products/dxp-cloud`
}
