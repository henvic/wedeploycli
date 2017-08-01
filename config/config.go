package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"gopkg.in/ini.v1"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/remotes"
	"github.com/wedeploy/cli/verbose"
)

// ContextType structure
type ContextType struct {
	Remote               string
	Infrastructure       string
	InfrastructureDomain string
	ServiceDomain        string
	Username             string
	Password             string
	Token                string
}

// Config of the application
type Config struct {
	DefaultRemote   string       `ini:"default_remote"`
	LocalHTTPPort   int          `ini:"local_http_port"`
	LocalHTTPSPort  int          `ini:"local_https_port"`
	NoAutocomplete  bool         `ini:"disable_autocomplete_autoinstall"`
	NoColor         bool         `ini:"disable_colors"`
	NotifyUpdates   bool         `ini:"notify_updates"`
	ReleaseChannel  string       `ini:"release_channel"`
	LastUpdateCheck string       `ini:"last_update_check"`
	PastVersion     string       `ini:"past_version"`
	NextVersion     string       `ini:"next_version"`
	EnableAnalytics bool         `ini:"enable_analytics"`
	AnalyticsID     string       `ini:"analytics_id"`
	Path            string       `ini:"-"`
	Remotes         remotes.List `ini:"-"`
	file            *ini.File    `ini:"-"`
	saveLocalRemote bool         // use custom local Remote
}

var (
	// Global configuration
	Global *Config

	// Context stores the environmental context
	Context *ContextType

	parseRemoteSectionNameRegex = regexp.MustCompile(`remote \"(.*)\"`)
)

// Load the configuration
func (c *Config) Load() error {
	if err := c.openOrCreate(); err != nil {
		return err
	}

	c.load()

	if c.Remotes == nil {
		c.Remotes = remotes.List{}
	}

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
	switch c.Remotes[defaults.CloudRemote].Infrastructure {
	case defaults.Infrastructure:
	case "":
		c.Remotes.Set(defaults.CloudRemote, remotes.Entry{
			Infrastructure:        defaults.Infrastructure,
			InfrastructureComment: "Default cloud remote",
		})
	default:
		println(color.Format(color.FgHiRed, "Warning: Non-standard wedeploy remote cloud detected"))
	}

	switch c.Remotes[defaults.LocalRemote].Infrastructure {
	case "":
		c.Remotes.Set(defaults.LocalRemote, remotes.Entry{
			Infrastructure:        c.getLocalEndpoint(),
			InfrastructureComment: "Default local remote",
			Service:               defaults.LocalServiceDomain,
			Username:              "no-reply@wedeploy.com",
			Password:              "cli-tool-password",
		})
	default:
		println(color.Format(color.FgHiRed, "Warning: Custom local remote detected"))
		c.saveLocalRemote = true
	}
}

func (c *Config) validateDefaultRemote() error {
	var keys = c.Remotes.Keys()

	for _, k := range keys {
		if c.DefaultRemote == k {
			return nil
		}
	}

	return fmt.Errorf(`Remote "%v" is set as default, but not found.
Please fix your ~/.we file`, c.DefaultRemote)
}

// Save the configuration
func (c *Config) Save() error {
	var cfg = c.file
	var err = cfg.ReflectFrom(c)

	if err != nil {
		return errwrap.Wrapf("Can not load configuration: {{err}}", err)
	}

	c.updateRemotes()
	c.simplify()

	err = cfg.SaveToIndent(c.Path, "    ")

	if err != nil {
		return errwrap.Wrapf("Can not save configuration: {{err}}", err)
	}

	return nil
}

// Setup the environment
func Setup(path string) (err error) {
	path, err = filepath.Abs(path)

	if err != nil {
		return err
	}

	Global = &Config{
		Path: path,
	}

	if err = Global.Load(); err != nil {
		verbose.Debug("Error setting up global config")
		return err
	}

	Context = &ContextType{}
	return nil
}

// SetEndpointContext for a given remote
func SetEndpointContext(remote string) error {
	var r, ok = Global.Remotes[remote]

	if !ok {
		return fmt.Errorf(`Error loading selected remote "%v"`, remote)
	}

	Context.Remote = remote

	switch {
	case strings.HasPrefix(r.Infrastructure, "http://localhost"):
		Context.Infrastructure = Global.getLocalEndpoint()
	default:
		Context.Infrastructure = "https://api." + r.Infrastructure
	}

	Context.InfrastructureDomain = getRemoteAddress(r.Infrastructure)
	Context.ServiceDomain = r.Service
	Context.Username = r.Username
	Context.Password = r.Password
	Context.Token = r.Token
	return nil
}

func getRemoteAddress(address string) string {
	var removePrefixes = []string{
		"http://",
		"https://",
	}

	for _, prefix := range removePrefixes {
		if strings.HasPrefix(address, prefix) {
			address = strings.TrimPrefix(address, prefix)
			break
		}
	}

	var h, _, err = net.SplitHostPort(address)

	if err != nil {
		return address
	}

	return h
}

func (c *Config) setDefaults() {
	c.EnableAnalytics = true
	c.LocalHTTPPort = defaults.LocalHTTPPort
	c.LocalHTTPSPort = defaults.LocalHTTPSPort
	c.NotifyUpdates = true
	c.ReleaseChannel = defaults.StableReleaseChannel
	c.DefaultRemote = defaults.CloudRemote

	// By design Windows users should see no color unless they enable it
	// Issue #51.
	// https://github.com/wedeploy/cli/issues/51
	if runtime.GOOS == "windows" {
		c.NoColor = true
	}
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

func (c *Config) read() error {
	var err error
	c.file, err = ini.Load(c.Path)

	if err != nil {
		return errwrap.Wrapf("Error reading configuration file: {{err}}\n"+
			"Fix "+c.Path+" by hand or erase it.", err)
	}

	c.readRemotes()
	return nil
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
	c.Remotes = remotes.List{}

	for _, k := range c.listRemotes() {
		remote := c.getRemote(k)
		infrastructure := remote.Key("infrastructure")
		service := remote.Key("service")
		username := remote.Key("username")
		password := remote.Key("password")
		token := remote.Key("token")
		comment := remote.Comment

		c.Remotes[k] = remotes.Entry{
			Comment:               strings.TrimSpace(comment),
			Infrastructure:        strings.TrimSpace(infrastructure.String()),
			InfrastructureComment: strings.TrimSpace(infrastructure.Comment),
			Service:               strings.TrimSpace(service.String()),
			ServiceComment:        strings.TrimSpace(service.Comment),
			Username:              strings.TrimSpace(username.String()),
			UsernameComment:       strings.TrimSpace(username.Comment),
			Password:              strings.TrimSpace(password.String()),
			PasswordComment:       strings.TrimSpace(password.Comment),
			Token:                 strings.TrimSpace(token.String()),
			TokenComment:          strings.TrimSpace(token.Comment),
		}
	}
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

		var omitempty = []string{"username", "password", "token"}

		for _, k := range omitempty {
			var key = remote.Key(k)
			if key.Value() == "" && key.Comment == "" {
				remote.DeleteKey(k)
			}
		}
	}
}

func (c *Config) updateRemotes() {
	for _, k := range c.listRemotes() {
		if _, ok := c.Remotes[k]; !ok {
			c.deleteRemote(k)
		}
	}

	for k, v := range c.Remotes {
		if !c.saveLocalRemote && k == defaults.LocalRemote {
			continue
		}

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

		keyPassword := remote.Key("password")
		keyPassword.SetValue(v.Password)
		keyPassword.Comment = v.PasswordComment

		keyToken := remote.Key("token")
		keyToken.SetValue(v.Token)
		keyToken.Comment = v.TokenComment

		remote.Comment = v.Comment
	}

	c.simplifyRemotes()
}

func (c *Config) banner() {
	c.file.Section("DEFAULT").Comment = `; Configuration file for WeDeploy CLI
; https://wedeploy.com`
}

func (c *Config) getLocalEndpoint() string {
	var endpoint = "http://localhost"

	if c.LocalHTTPPort != 80 {
		endpoint += fmt.Sprintf(":%d", c.LocalHTTPPort)
	}

	return endpoint
}

// Teardown resets the configuration environment
func Teardown() {
	Global = nil
	Context = nil
}
