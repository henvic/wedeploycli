package cmdremote

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/verbose"
)

// RemoteCmd runs the WeDeploy structure for development locally
var RemoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Configure WeDeploy remotes",
	RunE:  remoteRun,
}

var addCmd = &cobra.Command{
	Use:     "add",
	Short:   "Adds a remote named <name> for the repository at <url>",
	Example: "we remote add hk https://hk.example.com/",
	RunE:    addRun,
}

var renameCmd = &cobra.Command{
	Use:     "rename",
	Short:   "Rename the remote named <old> to <new>",
	Example: "we remote rename asia hk",
	RunE:    renameRun,
}

var removeCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove the remote named <name>",
	Example: "we remote remove hk",
	RunE:    removeRun,
}

var getURLCmd = &cobra.Command{
	Use:   "get-url",
	Short: "Retrieves the URLs for a remote",
	RunE:  getURLRun,
}

var setURLCmd = &cobra.Command{
	Use:   "set-url",
	Short: "Changes URLs for the remote",
	RunE:  setURLRun,
}

func remoteRun(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return errors.New("Invalid number of arguments.")
	}

	var remotes = config.Global.Remotes

	for _, k := range remotes.Keys() {
		switch verbose.Enabled {
		case true:
			var key, _ = remotes[k]
			fmt.Printf("%s\t%s\n", k, key.URL)
		default:
			fmt.Println(k)
		}
	}

	return nil
}

func addRun(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return errors.New("Invalid number of arguments.")
	}

	var global = config.Global
	var remotes = global.Remotes
	var name = args[0]

	if _, ok := remotes[name]; ok {
		return errors.New("fatal: remote " + name + " already exists.")
	}

	remotes.Set(name, args[1])
	return global.Save()
}

func renameRun(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return errors.New("Invalid number of arguments.")
	}

	var global = config.Global
	var remotes = global.Remotes
	var old = args[0]
	var name = args[1]

	var oldRemote, ok = remotes[old]

	if !ok {
		return errors.New("fatal: remote " + old + " does not exists.")
	}

	if _, ok := remotes[name]; ok {
		return errors.New("fatal: remote " + name + " already exists.")
	}

	remotes.Set(name, oldRemote.URL, oldRemote.Comment)
	remotes.Del(old)
	return global.Save()
}

func removeRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("This command takes 1 argument.")
	}

	var global = config.Global
	var remotes = global.Remotes
	var name = args[0]

	if _, ok := remotes[name]; !ok {
		return errors.New("fatal: remote " + name + " does not exists.")
	}

	remotes.Del(name)
	return global.Save()
}

func getURLRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("This command takes 1 argument.")
	}

	var remotes = config.Global.Remotes
	var name = args[0]
	var remote, ok = remotes[name]

	if !ok {
		return errors.New("fatal: remote " + name + " does not exists.")
	}

	fmt.Println(remote.URL)
	return nil
}

func setURLRun(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return errors.New("This command takes 2 arguments.")
	}

	var global = config.Global
	var remotes = global.Remotes
	var name = args[0]
	var uri = args[1]

	if _, ok := remotes[name]; !ok {
		return errors.New("fatal: remote " + name + " does not exists.")
	}

	remotes.Set(name, uri)
	return global.Save()
}

func init() {
	RemoteCmd.AddCommand(addCmd)
	RemoteCmd.AddCommand(renameCmd)
	RemoteCmd.AddCommand(removeCmd)
	RemoteCmd.AddCommand(getURLCmd)
	RemoteCmd.AddCommand(setURLCmd)
}
