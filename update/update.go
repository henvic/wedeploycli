package update

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/equinox-io/equinox"
	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/verbose"
)

var cacheNonAvailabilityHours = 12

// AppID is Equinox's app ID for this tool
var AppID = "app_g12mjgk2k9D"

// PublicKey is the public key for the certificate used with Equinox
var PublicKey = []byte(`
-----BEGIN ECDSA PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAE1+wgAlkBJRmtRwmZWfq9fa8dBlJ929hM
BzASHioHo6RP1V+4EKnAxaYXN4eWlgalxQ2BEr8TqYRM+uHPizteVR11wKfsO6S0
GENiOKpfivw5FIiTN14MZeMTagiKJUOq
-----END ECDSA PUBLIC KEY-----
`)

// GetReleaseChannel gets the channel user has used last time (or stable)
func GetReleaseChannel(c *config.Config) string {
	var channel = c.ReleaseChannel

	if channel == "" {
		channel = defaults.StableReleaseChannel
	}

	return channel
}

func canVerify(c *config.Config) bool {
	var next = c.NextVersion

	// is there an update being rolled at this exec time?
	if next != "" && next != defaults.Version {
		return false
	}

	// how long since last non availability result?
	return canVerifyAgain(c)
}

func canVerifyAgain(c *config.Config) bool {
	var luc, luce = time.Parse(time.RubyDate, c.LastUpdateCheck)

	if luce == nil && time.Since(luc).Hours() < float64(cacheNonAvailabilityHours) {
		return false
	}

	return true
}

// NotifierCheck enquires equinox if a new version is available
func NotifierCheck(c *config.Config) error {
	if !isNotifyOn(c) || !canVerify(c) {
		return nil
	}

	err := notifierCheck(c)

	switch err.(type) {
	case *url.Error:
		// Don't show connection error as the user might be off-line for a while
		if !verbose.Enabled && err != nil && strings.Contains(err.Error(), "no such host") {
			return nil
		}
	}

	return err
}

func notifierCheck(c *config.Config) error {
	// save, just to be safe (e.g., if the check below breaks)
	c.LastUpdateCheck = getCurrentTime()

	if err := c.Save(); err != nil {
		return err
	}

	var resp, err = check(GetReleaseChannel(c))

	switch err {
	case nil:
		c.NextVersion = resp.ReleaseVersion
		if err := c.Save(); err != nil {
			return err
		}
	case equinox.NotAvailableErr:
		c.NextVersion = ""
		if err := c.Save(); err != nil {
			return err
		}
		return nil
	}

	return err
}

// Notify is called every time this tool executes to verify if it is outdated
func Notify(c *config.Config) {
	if !isNotifyOn(c) {
		return
	}

	var next = c.NextVersion
	if next != "" && next != defaults.Version {
		notify(c)
	}
}

// Update this tool
func Update(c *config.Config, channel string) error {
	fmt.Printf("Current installed version is %s.\n", defaults.Version)
	fmt.Printf("Channel \"%s\" is now selected.\n", channel)

	var resp, err = check(channel)

	if err != nil {
		return handleUpdateCheckError(c, channel, err)
	}

	err = updateApply(c, channel, resp)

	if err != nil {
		return err
	}

	fmt.Printf("\nUpdated to version %s\n", resp.ReleaseVersion)
	runUpdateNotices()
	return nil
}

func runUpdateNotices() {
	var params = []string{"update", "release-notes", "--from", defaults.Version, "--exclusive"}
	verbose.Debug(fmt.Sprintf("Running %v %v", os.Args[0], strings.Join(params, " ")))
	cmd := exec.CommandContext(context.Background(), os.Args[0], params...)

	buf := new(bytes.Buffer)

	if verbose.Enabled {
		cmd.Stderr = os.Stderr
	}

	cmd.Stdout = buf

	if err := cmd.Run(); err != nil {
		verbose.Debug("we update release-notes error:", err)
		return
	}

	if buf.Len() != 0 {
		fmt.Print("\n\nChanges (release notes)\n")
		fmt.Print(buf)
	}
}

func check(channel string) (*equinox.Response, error) {
	var opts equinox.Options
	opts.Channel = channel

	if err := opts.SetPublicKeyPEM(PublicKey); err != nil {
		return nil, err
	}

	resp, err := equinox.Check(AppID, opts)

	return &resp, err
}

func getCurrentTime() string {
	return time.Now().Format(time.RubyDate)
}

func handleUpdateCheckError(c *config.Config, channel string, err error) error {
	switch {
	case err == equinox.NotAvailableErr:
		c.NextVersion = ""
		c.ReleaseChannel = channel
		c.LastUpdateCheck = getCurrentTime()
		if err := c.Save(); err != nil {
			return err
		}
		fmt.Println(fancy.Info("No update available."))
		return nil
	case strings.Contains(err.Error(), fmt.Sprintf("No channel with the name '%s' can be found.", channel)):
		return errwrap.Wrapf(fmt.Sprintf(`channel "%s" was not found`, channel), err)
	default:
		return errwrap.Wrapf("update failed: {{err}}", err)
	}
}

func isNotifyOn(c *config.Config) bool {
	if len(os.Args) == 2 && os.Args[1] == "update" {
		return false
	}

	if defaults.Version == "master" || !c.NotifyUpdates {
		return false
	}

	return true
}

func notify(c *config.Config) {
	var channel = c.ReleaseChannel
	var cmd = "we update"

	if channel != c.ReleaseChannel {
		cmd += " --channel " + channel
	}

	println(color.Format(color.FgBlue,
		`INFO: New version of WeDeploy CLI is available. Please run "%v".`,
		cmd))
}

func updateApply(c *config.Config, channel string, resp *equinox.Response) error {
	if err := resp.Apply(); err != nil {
		return handleUpdateApplyError(err)
	}

	return updateConfig(c, channel)
}

func handleUpdateApplyError(err error) error {
	if err != nil && strings.Contains(err.Error(), "permission denied") {
		return errwrap.Wrapf(`permission denied. Try "sudo we update" instead`, err)
	}

	return err
}

func updateConfig(c *config.Config, channel string) error {
	c.ReleaseChannel = channel
	c.PastVersion = defaults.Version
	c.NextVersion = ""
	c.LastUpdateCheck = getCurrentTime()
	return c.Save()
}
