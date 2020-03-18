// installer with reexecute functionality:
// after downloading and installing newest version, it reexecutes the command
package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/equinox-io/equinox"
	"github.com/henvic/uilive"
	"github.com/henvic/wedeploycli/update/keys"
)

// EnvSkipReexec environment variable option for skipping to re-execute command
const EnvSkipReexec = "WEDEPLOY_INSTALLER_SKIP_REEXEC"

// EnvVerbose option to print verbose info about the program
const EnvVerbose = "VERBOSE"

var (
	// Version of the Liferay Cloud CLI tool
	Version = "installer"

	// Build commit
	Build = ""

	// BuildTime is the time when the build was generated
	BuildTime = ""
)

type installation struct {
	ws *uilive.Writer
}

func check(channel string) (*equinox.Response, error) {
	var opts equinox.Options
	opts.Channel = channel

	if err := opts.SetPublicKeyPEM(keys.PublicKey); err != nil {
		return nil, err
	}

	resp, err := equinox.Check(keys.AppID, opts)

	return &resp, err
}

func (i *installation) install() error {
	var resp, err = check("stable")

	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(i.ws, "Downloading newest Liferay Cloud Platform CLI. Please wait.")
	_ = i.ws.Flush()
	return resp.Apply()
}

func (i *installation) run() error {
	i.ws = uilive.New()
	_, _ = fmt.Fprintln(i.ws, "Installing Liferay Cloud Platform CLI for the first time. Please wait.")
	_ = i.ws.Flush()

	if err := i.install(); err != nil {
		return err
	}

	_, _ = fmt.Fprint(i.ws, "Liferay Cloud Platform CLI tool installed. Thank you.\n\n")
	_ = i.ws.Flush()
	return nil
}

func reexecute() {
	var cmd = exec.Command(os.Args[0], os.Args[1:]...) // #nosec
	cmd.Env = os.Environ()
	cmd.Dir, _ = os.Getwd()
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()

	if err == nil {
		return
	}

	if exit, ok := err.(*exec.ExitError); ok {
		if process, ok := exit.Sys().(syscall.WaitStatus); ok {
			os.Exit(process.ExitStatus())
		}
	}

	_, _ = fmt.Fprintln(os.Stderr, "An unexpected error happened, please report it to https://help.liferay.com/hc/en-us/categories/360000813091-Liferay-DXP-Cloud-Documentation")
	panic(err)
}

func main() {
	if _, ok := os.LookupEnv(EnvVerbose); ok {
		printInstallerVersion()
	}

	var i = installation{}

	if err := i.run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if _, ok := os.LookupEnv(EnvSkipReexec); !ok {
		reexecute()
	}
}

func printInstallerVersion() {
	var os = runtime.GOOS
	var arch = runtime.GOARCH
	fmt.Printf("Liferay Cloud Platform CLI Installer version %s %s/%s\n",
		Version,
		os,
		arch)

	if Build != "" {
		fmt.Printf("Liferay Cloud Platform CLI Installer Build commit: %v\n", Build)
	}

	if BuildTime != "" {
		fmt.Printf("Liferay Cloud Platform CLI Installer Build time: %v\n", BuildTime)
	}

	fmt.Printf("Liferay Cloud Platform CLI Installer Go version: %s\n", runtime.Version())
}
