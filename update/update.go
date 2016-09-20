package update

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/equinox-io/equinox"
	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/run"
	"github.com/wedeploy/cli/verbose"
)

const lucFormat = "Mon Jan _2 15:04:05 MST 2006"

var cacheNonAvailabilityDays = 1

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
func GetReleaseChannel() string {
	var channel = config.Global.ReleaseChannel

	if channel == "" {
		channel = "stable"
	}

	return channel
}

func canVerify() bool {
	var next = config.Global.NextVersion

	// is there an update being rolled at this exec time?
	if next != "" && next != defaults.Version {
		return false
	}

	// how long since last non availability result?
	return canVerifyAgain()
}

func canVerifyAgain() bool {
	var luc, luce = time.Parse(lucFormat, config.Global.LastUpdateCheck)

	if luce == nil && time.Since(luc).Hours() < float64(cacheNonAvailabilityDays*24) {
		return false
	}

	return true
}

// NotifierCheck enquires equinox if a new version is available
func NotifierCheck() error {
	if !isNotifyOn() || !canVerify() {
		return nil
	}

	err := notifierCheck()

	switch err.(type) {
	case *url.Error:
		// Don't show connection error as the user might be off-line for a while
		if !verbose.Enabled && err != nil && strings.Contains(err.Error(), "no such host") {
			return nil
		}
	}

	return err
}

func notifierCheck() error {
	// save, just to be safe (e.g., if the check below breaks)
	var g = config.Global
	g.LastUpdateCheck = getCurrentTime()

	if err := g.Save(); err != nil {
		return err
	}

	var resp, err = check(GetReleaseChannel())

	switch err {
	case nil:
		g.NextVersion = resp.ReleaseVersion
		if err := g.Save(); err != nil {
			return err
		}
	case equinox.NotAvailableErr:
		g.NextVersion = ""
		if err := g.Save(); err != nil {
			return err
		}
		return nil
	}

	return err
}

// Notify is called every time this tool executes to verify if it is outdated
func Notify() {
	if !isNotifyOn() {
		return
	}

	var next = config.Global.NextVersion
	if next != "" && next != defaults.Version {
		notify()
	}
}

// Update this tool
func Update(channel string) error {
	fmt.Println("Trying to update using the", channel, "distribution channel")
	fmt.Println("Current installed version is " + defaults.Version)

	var resp, err = check(channel)

	if err != nil {
		return handleUpdateCheckError(err)
	}

	err = updateApply(channel, resp)

	if err != nil {
		return err
	}

	fmt.Println("Updated to new version:", resp.ReleaseVersion)
	return err
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
	return time.Now().Format(lucFormat)
}

func handleUpdateCheckError(err error) error {
	switch err {
	case equinox.NotAvailableErr:
		var g = config.Global
		g.NextVersion = ""
		g.LastUpdateCheck = getCurrentTime()
		if err := g.Save(); err != nil {
			return err
		}
		fmt.Println("No updates available.")
		return nil
	default:
		return errwrap.Wrapf("Update failed: {{err}}", err)
	}
}

func isNotifyOn() bool {
	if len(os.Args) == 2 && os.Args[1] == "update" {
		return false
	}

	if defaults.Version == "master" || !config.Global.NotifyUpdates {
		return false
	}

	return true
}

func notify() {
	var channel = config.Global.ReleaseChannel
	var cmd = "we update"

	if channel != "" && channel != "stable" {
		cmd += " --channel " + channel
	}

	println(color.Format(color.FgRed,
		`WARNING: WeDeploy CLI tool is outdated. Run "`+cmd+`".`))
}

func updateApply(channel string, resp *equinox.Response) error {
	if err := run.StopOutdatedImage(""); err != nil {
		return err
	}

	if err := resp.Apply(); err != nil {
		return err
	}

	return updateConfig(channel)
}

func updateConfig(channel string) error {
	var g = config.Global

	g.ReleaseChannel = channel
	g.PastVersion = defaults.Version
	g.NextVersion = ""
	g.LastUpdateCheck = getCurrentTime()
	return g.Save()
}
