package cmdremote

import (
	"errors"
	"fmt"
	"os"

	"strings"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/formatter"
	"github.com/wedeploy/cli/remotes"
)

// RemoteCmd runs the WeDeploy structure for development locally
var RemoteCmd = &cobra.Command{
	Use:     "remote",
	Short:   "Configure WeDeploy remotes",
	PreRunE: cmdargslen.ValidateCmd(0, 0),
	RunE:    remoteRun,
}

var setCmd = &cobra.Command{
	Use:     "set",
	Short:   "Set a remote named <name> with <url>",
	Aliases: []string{"add"},
	PreRunE: cmdargslen.ValidateCmd(2, 2),
	RunE:    setRun,
}

var renameCmd = &cobra.Command{
	Use:     "rename",
	Short:   "Rename the remote named <old> to <new>",
	PreRunE: cmdargslen.ValidateCmd(2, 2),
	RunE:    renameRun,
	Hidden:  true,
}

var removeCmd = &cobra.Command{
	Use:     "rm",
	Short:   "Remove the remote named <name>",
	PreRunE: cmdargslen.ValidateCmd(1, 1),
	RunE:    removeRun,
}

var getURLCmd = &cobra.Command{
	Use:     "get-url",
	Short:   "Retrieves the URLs for a remote",
	PreRunE: cmdargslen.ValidateCmd(1, 1),
	RunE:    getURLRun,
	Hidden:  true,
}

var setURLCmd = &cobra.Command{
	Use:         "set-url",
	Short:       "Changes URLs for the remote",
	PreRunE:     cmdargslen.ValidateCmd(2, 2),
	RunE:        setURLRun,
	Hidden:      true,
	Annotations: nil,
}

func remoteRun(cmd *cobra.Command, args []string) error {
	var remotes = config.Global.Remotes
	var w = formatter.NewTabWriter(os.Stdout)

	for _, k := range remotes.Keys() {
		var key, _ = remotes[k]
		var infrastructure = key.Infrastructure

		if k == defaults.LocalRemote {
			infrastructure = strings.TrimPrefix(infrastructure, "http://")
		}

		fmt.Fprintf(w, "%s\t%s", k, infrastructure)

		if k == config.Global.DefaultRemote {
			fmt.Fprintf(w, " (default)")
		}

		fmt.Fprintf(w, "\n")
	}

	_ = w.Flush()

	return nil
}

func setRun(cmd *cobra.Command, args []string) error {
	var global = config.Global
	var r = global.Remotes
	var name = args[0]

	if _, ok := r[name]; ok {
		r.Del(name)
	}

	r.Set(name, remotes.Entry{
		Infrastructure: args[1],
	})

	return global.Save()
}

func renameRun(cmd *cobra.Command, args []string) error {
	var global = config.Global
	var r = global.Remotes
	var old = args[0]
	var name = args[1]

	var oldRemote, ok = r[old]

	if !ok {
		return errors.New("fatal: remote " + old + " does not exists.")
	}

	if _, ok := r[name]; ok {
		return errors.New("fatal: remote " + name + " already exists.")
	}

	r.Del(old)
	r.Set(name, oldRemote)
	return global.Save()
}

func removeRun(cmd *cobra.Command, args []string) error {
	var global = config.Global
	var remotes = global.Remotes
	var name = args[0]

	if _, ok := remotes[name]; !ok {
		return errors.New("fatal: remote " + name + " does not exists.")
	}

	remotes.Del(name)

	if name == defaults.CloudRemote {
		fmt.Fprintf(os.Stderr, "%v\n", color.Format(color.FgHiRed, `Removed default cloud remote "wedeploy" will be recreated with its default value`))
	}

	if name == defaults.LocalRemote {
		fmt.Fprintf(os.Stderr, "%v\n", color.Format(color.FgHiRed, `Removed default local remote "local" will be recreated with its default value`))
	}

	if name == global.DefaultRemote && name != defaults.CloudRemote && name != defaults.LocalRemote {
		global.DefaultRemote = defaults.CloudRemote
		fmt.Fprintf(os.Stderr, "%v\n", color.Format(color.FgHiRed, `Default remote reset to "wedeploy"`))
	}

	return global.Save()
}

func getURLRun(cmd *cobra.Command, args []string) error {
	var remotes = config.Global.Remotes
	var name = args[0]
	var remote, ok = remotes[name]

	if !ok {
		return errors.New("fatal: remote " + name + " does not exists.")
	}

	fmt.Println(remote.Infrastructure)
	return nil
}

func setURLRun(cmd *cobra.Command, args []string) error {
	var global = config.Global
	var r = global.Remotes
	var name = args[0]
	var uri = args[1]

	if _, ok := r[name]; !ok {
		return errors.New("fatal: remote " + name + " does not exists.")
	}

	r.Set(name, remotes.Entry{
		Infrastructure: uri,
	})
	return global.Save()
}

func init() {
	RemoteCmd.AddCommand(setCmd)
	RemoteCmd.AddCommand(renameCmd)
	RemoteCmd.AddCommand(removeCmd)
	RemoteCmd.AddCommand(getURLCmd)
	RemoteCmd.AddCommand(setURLCmd)
}
