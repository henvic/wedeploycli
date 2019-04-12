package pkgrename

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/henvic/climetricsvendor/github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/userhome"
)

// Notice tells the user the tool was rebranded
func Notice() {
	fmt.Println(color.Format(color.FgBlue, "WeDeploy CLI is now Liferay CLI."))
	fmt.Println(`From now on use "liferay" instead of "we". Example: "liferay deploy"`)
	fmt.Print(color.Format(color.FgRed, "\nYou should convert your wedeploy.json files to wedeploy.json now.\n"))
}

// MoveToLiferay moves 'we' binary to 'liferay'
func MoveToLiferay() error {
	if os.Args[0] != "we" {
		return nil
	}

	var executable, err = os.Executable()

	if err != nil {
		return err
	}

	executable, err = filepath.EvalSymlinks(executable)

	if err != nil {
		return err
	}

	if filepath.Base(executable) != "we" {
		return nil
	}

	var newPath = filepath.Join(filepath.Dir(executable), "liferay")

	if _, err = os.Stat(newPath); err == nil {
		fmt.Fprintln(os.Stderr, `Bypassing moving "we" to "liferay" because "liferay" already exists."`)
		return nil
	}

	return os.Rename(executable, newPath)
}

var moved = []move{
	{from: ".wedeploy-deploys", to: ".liferay-deploys"},
	{from: ".we", to: ".liferaycli"},
	{from: ".we_autocomplete", to: ".liferaycli_autocomplete"},
	{from: ".we_metrics", to: ".liferaycli_metrics"},
}

// MoveConfiguration from $HOME/.we to $HOME/.liferay
func MoveConfiguration() error {
	var homeDir = userhome.GetHomeDir()

	if !hasOldConfiguration() {
		return nil
	}

	for _, m := range moved {
		err := os.Rename(filepath.Join(homeDir, m.from), filepath.Join(homeDir, m.to))

		if err != nil && !os.IsNotExist(err) {
			return errwrap.Wrapf("cannot move "+m.from+": {{err}}", err)
		}
	}

	fmt.Println(color.Format(color.FgRed, "Moved Liferay configuration to ~/.liferaycli"))
	return nil
}

func hasOldConfiguration() bool {
	for _, m := range moved {
		var homeDir = userhome.GetHomeDir()
		var _, err = os.Stat(filepath.Join(homeDir, m.from))

		if err == nil {
			return true
		}
	}

	return false
}

type move struct {
	from, to string
}
