package uninstall

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/userhome"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/waitlivemsg"
)

// UninstallCmd is used for uninstall this tool
var UninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Args:  cobra.NoArgs,
	RunE:  uninstallRun,
	Short: "Uninstall CLI",
}

var rmConfig bool

func uninstall() error {
	var exec, err = os.Executable()

	if err != nil {
		return err
	}

	return os.Remove(exec)
}

func uninstallChan(m *waitlivemsg.Message, ec chan error) {
	var err = removeConfig()

	if err != nil {
		m.StopText(fancy.Error("Failed to remove user profile, configuration, and uninstall CLI [1/2]"))
		ec <- err
		return
	}

	err = uninstall()

	if err != nil {
		m.StopText(fancy.Error("Failed to uninstall the CLI [1/2]"))
		ec <- err
		return
	}

	m.StopText("Liferay Cloud Platform CLI uninstalled [2/2]\n" +
		fancy.Info("Liferay Cloud Platform CLI is not working on your computer anymore.") + "\n" +
		color.Format(color.FgHiYellow, "  For installing it again, type this command and press Enter:\n") +
		color.Format(color.FgHiBlack, "  $ curl http://cdn.wedeploy.com/cli/latest/wedeploy.sh -sL | bash"))
	ec <- err
}

func removeConfigOnlyChan(m *waitlivemsg.Message, ec chan error) {
	err := removeConfig()

	if err != nil {
		m.StopText(fancy.Error("Failed to remove user profile and configuration [1/2]"))
		ec <- err
		return
	}

	m.StopText("User profile and configuration removed [2/2]")
	ec <- nil
}

func removeConfig() error {
	var homeDir = userhome.GetHomeDir()

	var files = []string{
		filepath.Join(homeDir, ".wedeploy"), // cache directory
		filepath.Join(homeDir, ".lcp"),
		filepath.Join(homeDir, ".lcp_autocomplete"),
		filepath.Join(homeDir, ".lcp_metrics"),
	}

	var el []string

	for _, f := range files {
		verbose.Debug("Removing " + f)
		err := os.RemoveAll(f)
		if err != nil {
			el = append(el, err.Error())
		}
	}

	if len(el) == 0 {
		return nil
	}

	return errors.New("can't remove all files: " + strings.Join(el, "\n"))
}

func uninstallRoutine(m *waitlivemsg.Message, ec chan error) {
	if rmConfig {
		removeConfigOnlyChan(m, ec)
		return
	}

	uninstallChan(m, ec)
}

func uninstallRun(cmd *cobra.Command, args []string) error {
	var m *waitlivemsg.Message
	switch rmConfig {
	case true:
		m = waitlivemsg.NewMessage("Removing configuration files [1/2]")
	default:
		if runtime.GOOS == "windows" {
			_, _ = fmt.Fprintln(os.Stderr, "Can't self-uninstall on Windows yet. Please remove the Liferay Cloud Platform CLI in the Control Panel.")
		}

		m = waitlivemsg.NewMessage("Uninstalling the Liferay Cloud Platform CLI [1/2]")
	}

	var wlm = waitlivemsg.New(nil)
	go wlm.Wait()
	wlm.AddMessage(m)
	var ec = make(chan error, 1)
	go uninstallRoutine(m, ec)
	var err = <-ec
	wlm.Stop()
	return err
}

func init() {
	UninstallCmd.Flags().BoolVar(&rmConfig, "rm-config", false, "Remove user profile and configuration only")
}
