package config

import (
	"fmt"
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
	"github.com/wedeploy/cli/remoteuriparser"
	"github.com/wedeploy/cli/usercontext"
	"github.com/wedeploy/cli/verbose"
)

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
	AnalyticsOption string       `ini:"analytics_option_date"`
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
	Context *usercontext.Context

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
	switch c.Remotes[defaults.CloudRemote].URL {
	case "wedeploy.io":
	case "":
		c.Remotes.Set(defaults.CloudRemote, remotes.Entry{
			URL:        "wedeploy.io",
			URLComment: "Default cloud remote",
		})
	default:
		println(color.Format(color.FgHiRed, "Warning: Non-standard wedeploy remote cloud detected"))
	}

	var (
		localRemoteURL       = c.Remotes[defaults.LocalRemote].URL
		currentLocalEndpoint = getLocalEndpoint()
	)

	switch localRemoteURL {
	case "":
		c.Remotes.Set(defaults.LocalRemote, remotes.Entry{
			URL:        currentLocalEndpoint,
			URLComment: "Default local remote",
			Username:   "no-reply@wedeploy.com",
			Password:   "cli-tool-password",
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

// IsEndpoint returns a boolean telling whether the URL is within a WeDeploy endpoint or not
func (c *Config) IsEndpoint(host string) bool {
	for _, remote := range c.Remotes {
		// this could be improved
		if len(remote.URL) != 0 && strings.Contains(remote.URL, host) {
			return true
		}
	}

	return false
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

	Context = &usercontext.Context{}
	return Context.Load()
}

// SetEndpointContext for a given remote
func SetEndpointContext(remote string) error {
	var r, ok = Global.Remotes[remote]

	if !ok {
		return fmt.Errorf(`Error loading selected remote "%v"`, remote)
	}

	Context.Remote = remote
	Context.RemoteAddress = getRemoteAddress(r.URL)
	Context.Endpoint = remoteuriparser.Parse(r.URL)
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
			return strings.TrimPrefix(address, prefix)
		}
	}

	return address
}

func (c *Config) setDefaults() {
	c.LocalHTTPPort = defaults.LocalHTTPPort
	c.LocalHTTPSPort = defaults.LocalHTTPSPort
	c.NotifyUpdates = true
	c.ReleaseChannel = "stable"
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
		s := c.getRemote(k)
		u := s.Key("url")
		username := s.Key("username")
		password := s.Key("password")
		token := s.Key("token")
		comment := s.Comment

		c.Remotes[k] = remotes.Entry{
			URL:             strings.TrimSpace(u.String()),
			URLComment:      strings.TrimSpace(u.Comment),
			Comment:         strings.TrimSpace(comment),
			Username:        strings.TrimSpace(username.String()),
			UsernameComment: strings.TrimSpace(username.Comment),
			Password:        strings.TrimSpace(password.String()),
			PasswordComment: strings.TrimSpace(password.Comment),
			Token:           strings.TrimSpace(token.String()),
			TokenComment:    strings.TrimSpace(token.Comment),
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
		s := c.getRemote(k)
		u := s.Key("url")

		if u.Value() == "" && u.Comment == "" {
			s.DeleteKey("url")
		}

		if len(s.Keys()) == 0 && s.Comment == "" {
			c.deleteRemote(k)
		}

		var omitempty = []string{"username", "password", "token"}

		for _, k := range omitempty {
			var key = s.Key(k)
			if key.Value() == "" && key.Comment == "" {
				s.DeleteKey(k)
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

		s := c.getRemote(k)
		keyURL := s.Key("url")
		keyURL.SetValue(v.URL)
		keyURL.Comment = v.URLComment

		keyUsername := s.Key("username")
		keyUsername.SetValue(v.Username)
		keyUsername.Comment = v.UsernameComment

		keyPassword := s.Key("password")
		keyPassword.SetValue(v.Password)
		keyPassword.Comment = v.PasswordComment

		keyToken := s.Key("token")
		keyToken.SetValue(v.Token)
		keyToken.Comment = v.TokenComment

		s.Comment = v.Comment
	}

	c.simplifyRemotes()
}

func (c *Config) banner() {
	c.file.Section("DEFAULT").Comment = `; Configuration file for WeDeploy CLI
; https://wedeploy.com`
}

func getLocalEndpoint() string {
	var endpoint = "http://wedeploy.me"

	if Global.LocalHTTPPort != 80 {
		endpoint += fmt.Sprintf(":%d", Global.LocalHTTPPort)
	}

	return endpoint
}

// Teardown resets the configuration environment
func Teardown() {
	Global = nil
	Context = nil
}
