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

// Context structure
type Context struct {
	config  *Config
	context *ContextParams
}

// NewContext with received params and uninitialized configuration
func NewContext(params ContextParams) Context {
	return Context{
		context: &ContextParams{},
		config:  &Config{},
	}
}

// ContextParams is the set of environment configurations
type ContextParams struct {
	Remote               string
	Infrastructure       string
	InfrastructureDomain string
	ServiceDomain        string
	Username             string
	Token                string
}

// Remote used on the context
func (c *Context) Remote() string {
	return c.context.Remote
}

// Infrastructure used on the context
func (c *Context) Infrastructure() string {
	return c.context.Infrastructure
}

// InfrastructureDomain used on the context
func (c *Context) InfrastructureDomain() string {
	return c.context.InfrastructureDomain
}

// ServiceDomain used on the context
func (c *Context) ServiceDomain() string {
	return c.context.ServiceDomain
}

// Username used on the context
func (c *Context) Username() string {
	return c.context.Username
}

// Token used on the context
func (c *Context) Token() string {
	return c.context.Token
}

// Config gets the configuration
func (c *Context) Config() *Config {
	return c.config
}

// SetEndpoint for the context
func (c *Context) SetEndpoint(remote string) error {
	var r, ok = c.Config().Remotes[remote]

	if !ok {
		return fmt.Errorf(`Error loading selected remote "%v"`, remote)
	}

	c.context.Remote = remote
	var address = r.Infrastructure

	if !isHTTPLocalhost(address) {
		address = "https://api." + address
	}

	c.context.Infrastructure = address

	c.context.InfrastructureDomain = getRemoteAddress(r.Infrastructure)
	c.context.ServiceDomain = r.Service
	c.context.Username = r.Username
	c.context.Token = r.Token
	return nil
}

// Config of the application
type Config struct {
	DefaultRemote   string       `ini:"default_remote"`
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
}

var parseRemoteSectionNameRegex = regexp.MustCompile(`remote \"(.*)\"`)

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
func Setup(path string) (wectx Context, err error) {
	wectx = NewContext(ContextParams{})
	path, err = filepath.Abs(path)

	if err != nil {
		return wectx, err
	}

	var c = &Config{
		Path: path,
	}

	if err = c.Load(); err != nil {
		return wectx, err
	}

	wectx.config = c
	return wectx, nil
}

func isHTTPLocalhost(address string) bool {
	address = strings.TrimPrefix(address, "http://")
	var h, _, err = net.SplitHostPort(address)

	if err != nil {
		return false
	}

	return h == "localhost"
}

func getRemoteAddress(address string) string {
	if strings.HasPrefix(address, "https://api.") {
		address = strings.TrimPrefix(address, "https://api.")
	}

	var h, _, err = net.SplitHostPort(address)

	if err != nil {
		return address
	}

	return h
}

func (c *Config) setDefaults() {
	c.EnableAnalytics = true
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
	c.file.Section("DEFAULT").Comment = `; Configuration file for WeDeploy CLI
; https://wedeploy.com`
}
