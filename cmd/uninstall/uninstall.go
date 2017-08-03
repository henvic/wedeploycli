package cmduninstall

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/userhome"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/waitlivemsg"
)

// UninstallCmd is used for uninstall this tool
var UninstallCmd = &cobra.Command{
	Use:     "uninstall",
	PreRunE: cmdargslen.ValidateCmd(0, 0),
	RunE:    uninstallRun,
	Short:   "Uninstall CLI",
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

	m.StopText(fancy.Success("WeDeploy CLI uninstalled [2/2]") + "\n" +
		fancy.Info("WeDeploy CLI is not working on your computer anymore.") + "\n" +
		color.Format(color.FgHiYellow, "  For installing it again, type this command and press Enter:\n") +
		color.Format(color.FgHiBlack, "  $ curl http://cdn.wedeploy.com/cli/latest/wedeploy.sh -sL | bash"))
	ec <- err
}

func removeConfigOnlyChan(m *waitlivemsg.Message, ec chan error) {
	if err := removeConfig(); err != nil {
		m.StopText(fancy.Error("Failed to remove user profile and configuration [1/2]"))
		ec <- err
	}
}

func removeConfig() error {
	var homeDir = userhome.GetHomeDir()

	var files = []string{
		filepath.Join(homeDir, ".we"),
		filepath.Join(homeDir, ".we_autocomplete"),
		filepath.Join(homeDir, ".we_metrics"),
	}

	for _, f := range files {
		verbose.Debug("Removing " + f)
		err := os.Remove(f)
		if err != nil && os.IsNotExist(err) {
			return err
		}
	}

	verbose.Debug("Removing ~/.wedeploy cache directory")
	return os.RemoveAll(filepath.Join(homeDir, ".wedeploy"))
}

func uninstallRoutine(m *waitlivemsg.Message, ec chan error) {
	if rmConfig {
		removeConfigOnlyChan(m, ec)
		return
	}

	uninstallChan(m, ec)
}

func uninstallRun(cmd *cobra.Command, args []string) error {
	var m = waitlivemsg.NewMessage("Uninstalling the WeDeploy CLI [1/2]")
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
