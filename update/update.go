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
	"github.com/henvic/wedeploycli/color"
	"github.com/henvic/wedeploycli/config"
	"github.com/henvic/wedeploycli/defaults"
	"github.com/henvic/wedeploycli/fancy"
	"github.com/henvic/wedeploycli/update/keys"
	"github.com/henvic/wedeploycli/verbose"
)

var cacheNonAvailabilityHours = 12

// GetReleaseChannel gets the channel user has used last time (or stable)
func GetReleaseChannel(c *config.Config) string {
	var params = c.GetParams()
	var channel = params.ReleaseChannel

	if channel == "" {
		channel = defaults.StableReleaseChannel
	}

	return channel
}

func canVerify(c *config.Config) bool {
	var params = c.GetParams()
	var next = params.NextVersion

	// is there an update being rolled at this exec time?
	if next != "" && next != defaults.Version {
		return false
	}

	// how long since last non availability result?
	return canVerifyAgain(c)
}

func canVerifyAgain(c *config.Config) bool {
	var params = c.GetParams()
	var luc, luce = time.Parse(time.RubyDate, params.LastUpdateCheck)

	if luce == nil && time.Since(luc).Hours() < float64(cacheNonAvailabilityHours) {
		return false
	}

	return true
}

// NotifierCheck enquires equinox if a new version is available
func NotifierCheck(ctx context.Context, c *config.Config) error {
	if !isNotifyOn(c) || !canVerify(c) {
		return nil
	}

	err := notifierCheck(ctx, c)

	switch err.(type) {
	case *url.Error:
		// Don't show connection error as the user might be off-line for a while
		if !verbose.Enabled && err != nil && strings.Contains(err.Error(), "no such host") {
			return nil
		}
	}

	return err
}

func notifierCheck(ctx context.Context, c *config.Config) error {
	var params = c.GetParams()
	params.LastUpdateCheck = getCurrentTime()

	var resp, err = check(ctx, GetReleaseChannel(c), "")

	switch {
	case err == equinox.NotAvailableErr:
		params.NextVersion = ""
		c.SetParams(params)
		return c.Save()
	case err != nil:
		return err
	default:
		params.NextVersion = resp.ReleaseVersion
		c.SetParams(params)
		return c.Save()
	}
}

// Notify is called every time this tool executes to verify if it is outdated
func Notify(c *config.Config) {
	if !isNotifyOn(c) {
		return
	}

	var params = c.GetParams()
	var next = params.NextVersion
	if next != "" && next != defaults.Version {
		notify(c)
	}
}

// Update this tool
func Update(ctx context.Context, c *config.Config, channel, version string) error {
	fmt.Printf("Current installed version is %s.\n", defaults.Version)
	fmt.Printf("Channel \"%s\" is now selected.\n", channel)

	var resp, err = check(ctx, channel, version)

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
	cmd := exec.CommandContext(context.Background(), os.Args[0], params...) // #nosec

	buf := new(bytes.Buffer)

	if verbose.Enabled {
		cmd.Stderr = os.Stderr
	}

	cmd.Stdout = buf

	if err := cmd.Run(); err != nil {
		verbose.Debug("lcp update release-notes error:", err)
		return
	}

	if buf.Len() != 0 {
		fmt.Print("\n\nChanges (release notes)\n")
		fmt.Print(buf)
	}
}

func check(ctx context.Context, channel, version string) (*equinox.Response, error) {
	var opts = equinox.Options{
		CurrentVersion: defaults.Version,
		Channel:        channel,
		Version:        version,
	}

	if err := opts.SetPublicKeyPEM(keys.PublicKey); err != nil {
		return nil, err
	}

	resp, err := equinox.CheckContext(ctx, keys.AppID, opts)

	return &resp, err
}

func getCurrentTime() string {
	return time.Now().Format(time.RubyDate)
}

func handleUpdateCheckError(c *config.Config, channel string, err error) error {
	if err == equinox.NotAvailableErr {
		params := c.GetParams()

		params.NextVersion = ""
		params.ReleaseChannel = channel
		params.LastUpdateCheck = getCurrentTime()

		c.SetParams(params)

		if err = c.Save(); err != nil {
			return err
		}

		fmt.Println(fancy.Info("No update available."))
		return nil
	}

	if strings.Contains(err.Error(),
		fmt.Sprintf("No channel with the name '%s' can be found.", channel)) {
		return errwrap.Wrapf(fmt.Sprintf(`channel "%s" was not found`, channel), err)
	}

	return errwrap.Wrapf("update failed: {{err}}", err)
}

func isNotifyOn(c *config.Config) bool {
	if len(os.Args) == 2 && os.Args[1] == "update" {
		return false
	}

	var params = c.GetParams()

	if defaults.Version == "master" || !params.NotifyUpdates {
		return false
	}

	return true
}

func notify(c *config.Config) {
	var params = c.GetParams()
	var channel = params.ReleaseChannel
	var cmd = "lcp update"

	if channel != params.ReleaseChannel {
		cmd += " --channel " + channel
	}

	_, _ = fmt.Fprintln(os.Stderr, color.Format(color.FgBlue,
		`
INFO: New version of Liferay Cloud Platform CLI is available. Please run "%v".`,
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
		return errwrap.Wrapf(`permission denied. Try "sudo lcp update" instead`, err)
	}

	return err
}

func updateConfig(c *config.Config, channel string) error {
	var params = c.GetParams()

	params.ReleaseChannel = channel
	params.PastVersion = defaults.Version
	params.NextVersion = ""
	params.LastUpdateCheck = getCurrentTime()

	c.SetParams(params)
	return c.Save()
}
