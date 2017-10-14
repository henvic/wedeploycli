package cmdremote

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/formatter"
	"github.com/wedeploy/cli/remotes"
)

// RemoteCmd runs the WeDeploy structure for development locally
var RemoteCmd = &cobra.Command{
	Use:     "remote",
	Hidden:  true,
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
	var wectx = we.Context()
	var conf = wectx.Config()
	var remotes = conf.Remotes
	var w = formatter.NewTabWriter(os.Stdout)

	for _, k := range remotes.Keys() {
		var key, _ = remotes[k]
		var infrastructure = key.Infrastructure

		fmt.Fprintf(w, "%s\t%s", k, infrastructure)

		if k == conf.DefaultRemote {
			fmt.Fprintf(w, " (default)")
		}

		fmt.Fprintf(w, "\n")
	}

	_ = w.Flush()

	return nil
}

func setRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()
	var conf = wectx.Config()
	var r = conf.Remotes
	var name = args[0]

	if _, ok := r[name]; ok {
		r.Del(name)
	}

	r.Set(name, remotes.Entry{
		Infrastructure: args[1],
	})

	return conf.Save()
}

func renameRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()
	var conf = wectx.Config()
	var r = conf.Remotes
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
	return conf.Save()
}

func removeRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()
	var conf = wectx.Config()
	var remotes = conf.Remotes
	var name = args[0]

	if _, ok := remotes[name]; !ok {
		return errors.New("fatal: remote " + name + " does not exists.")
	}

	remotes.Del(name)

	if name == defaults.CloudRemote {
		fmt.Fprintf(os.Stderr, "%v\n", color.Format(color.FgHiRed, `Removed default cloud remote "wedeploy" will be recreated with its default value`))
	}

	if name == conf.DefaultRemote && name != defaults.CloudRemote {
		conf.DefaultRemote = defaults.CloudRemote
		fmt.Fprintf(os.Stderr, "%v\n", color.Format(color.FgHiRed, `Default remote reset to "wedeploy"`))
	}

	return conf.Save()
}

func getURLRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()
	var conf = wectx.Config()
	var remotes = conf.Remotes
	var name = args[0]
	var remote, ok = remotes[name]

	if !ok {
		return errors.New("fatal: remote " + name + " does not exists.")
	}

	fmt.Println(remote.Infrastructure)
	return nil
}

func setURLRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()
	var conf = wectx.Config()
	var r = conf.Remotes
	var name = args[0]
	var uri = args[1]

	if _, ok := r[name]; !ok {
		return errors.New("fatal: remote " + name + " does not exists.")
	}

	r.Set(name, remotes.Entry{
		Infrastructure: uri,
	})
	return conf.Save()
}

func init() {
	RemoteCmd.AddCommand(setCmd)
	RemoteCmd.AddCommand(renameCmd)
	RemoteCmd.AddCommand(removeCmd)
	RemoteCmd.AddCommand(getURLCmd)
	RemoteCmd.AddCommand(setURLCmd)
}
