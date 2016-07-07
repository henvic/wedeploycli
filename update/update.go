package update

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/equinox-io/equinox"
	"github.com/fatih/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
)

const lucFormat = "Mon Jan _2 15:04:05 MST 2006"

var cacheNonAvailabilityDays = 4

// AppID is Equinox's app ID for this tool
var AppID = "app_g12mjgk2k9D"

// ErrMasterVersion triggered when trying to update a development version
var ErrMasterVersion = errors.New(
	"You must update a development version manually with git")

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
	switch {
	case isNotifyOn() && canVerify():
		return notifierCheck()
	default:
		return nil
	}
}

func notifierCheck() error {
	// save, just to be safe (e.g., if the check below breaks)
	var g = config.Global
	g.LastUpdateCheck = getCurrentTime()
	g.Save()

	var resp, err = check(GetReleaseChannel())

	switch err {
	case nil:
		g.NextVersion = resp.ReleaseVersion
		g.Save()
	case equinox.NotAvailableErr:
		g.NextVersion = ""
		g.Save()
		err = nil
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
func Update(channel string) {
	if defaults.Version == "master" {
		fmt.Fprintln(os.Stderr, ErrMasterVersion)
		os.Exit(1)
	}

	fmt.Println("Trying to update using the", channel, "distribution channel")
	fmt.Println("Current installed version is " + defaults.Version)

	var resp, err = check(channel)
	handleUpdateCheckError(err)
	updateApply(channel, resp)

	fmt.Println("Updated to new version:", resp.ReleaseVersion)
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

func handleUpdateCheckError(err error) {
	switch err {
	case nil:
	case equinox.NotAvailableErr:
		var g = config.Global
		g.NextVersion = ""
		g.LastUpdateCheck = getCurrentTime()
		g.Save()
		fmt.Println("No updates available.")
		return
	default:
		println("Update failed:", err.Error())
		os.Exit(1)
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

	println(color.RedString(
		`WARNING: WeDeploy CLI tool is outdated. Run "` + cmd + `".`))
}

func updateApply(channel string, resp *equinox.Response) {
	var err = resp.Apply()

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	updateConfig(channel)
}

func updateConfig(channel string) {
	var g = config.Global

	g.ReleaseChannel = channel
	g.NextVersion = ""
	g.LastUpdateCheck = getCurrentTime()
	g.Save()
}
