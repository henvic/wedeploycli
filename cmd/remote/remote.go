package remote

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/formatter"
	"github.com/wedeploy/cli/remotes"
)

// RemoteCmd is the command to control WeDeploy remotes.
var RemoteCmd = &cobra.Command{
	Use:    "remote",
	Hidden: true,
	Short:  "Configure WeDeploy remotes",
	Args:   cobra.NoArgs,
	RunE:   remoteRun,
}

var setCmd = &cobra.Command{
	Use:     "set",
	Short:   "Set a remote named <name> with <url>",
	Aliases: []string{"add"},
	Args:    cobra.ExactArgs(2),
	RunE:    setRun,
}

var renameCmd = &cobra.Command{
	Use:    "rename",
	Short:  "Rename the remote named <old> to <new>",
	Args:   cobra.ExactArgs(2),
	RunE:   renameRun,
	Hidden: true,
}

var removeCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove the remote named <name>",
	Args:  cobra.ExactArgs(1),
	RunE:  removeRun,
}

var getURLCmd = &cobra.Command{
	Use:    "get-url",
	Short:  "Retrieves the URLs for a remote",
	Args:   cobra.ExactArgs(1),
	RunE:   getURLRun,
	Hidden: true,
}

var setURLCmd = &cobra.Command{
	Use:         "set-url",
	Short:       "Changes URLs for the remote",
	Args:        cobra.ExactArgs(2),
	RunE:        setURLRun,
	Hidden:      true,
	Annotations: nil,
}

func remoteRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()
	var conf = wectx.Config()
	var params = conf.GetParams()
	var rl = params.Remotes

	var w = formatter.NewTabWriter(os.Stdout)

	for _, k := range rl.Keys() {
		var key = rl.Get(k)
		var infrastructure = key.Infrastructure

		_, _ = fmt.Fprintf(w, "%s\t%s", k, infrastructure)

		if k == params.DefaultRemote {
			_, _ = fmt.Fprintf(w, " (default)")
		}

		_, _ = fmt.Fprintf(w, "\n")
	}

	_ = w.Flush()

	return nil
}

func setRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()
	var conf = wectx.Config()
	var params = conf.GetParams()
	var r = params.Remotes

	var name = args[0]

	if r.Has(name) {
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
	var params = conf.GetParams()
	var r = params.Remotes

	var old = args[0]
	var name = args[1]

	if !r.Has(old) {
		return errors.New("fatal: remote " + old + " does not exists.")
	}

	if r.Has(name) {
		return errors.New("fatal: remote " + name + " already exists.")
	}

	var renamed = r.Get(old)

	r.Del(old)
	r.Set(name, renamed)
	return conf.Save()
}

func removeRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()
	var conf = wectx.Config()
	var params = conf.GetParams()
	var rl = params.Remotes

	var name = args[0]

	if !rl.Has(name) {
		return errors.New("fatal: remote " + name + " does not exists.")
	}

	rl.Del(name)

	if name == defaults.CloudRemote {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", color.Format(color.FgHiRed, `Removed default cloud remote "wedeploy" will be recreated with its default value`))
	}

	if name == params.DefaultRemote && name != defaults.CloudRemote {
		params.DefaultRemote = defaults.CloudRemote
		conf.SetParams(params)
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", color.Format(color.FgHiRed, `Default remote reset to "wedeploy"`))
	}

	return conf.Save()
}

func getURLRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()
	var conf = wectx.Config()
	var params = conf.GetParams()
	var rl = params.Remotes

	var name = args[0]

	if !rl.Has(name) {
		return errors.New("fatal: remote " + name + " does not exists.")
	}

	var remote = rl.Get(name)

	fmt.Println(remote.Infrastructure)
	return nil
}

func setURLRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()
	var conf = wectx.Config()
	var params = conf.GetParams()
	var rl = params.Remotes

	var name = args[0]
	var uri = args[1]

	if !rl.Has(name) {
		return errors.New("fatal: remote " + name + " does not exists.")
	}

	var remote = rl.Get(name)

	remote.Infrastructure = uri

	rl.Set(name, remote)

	return conf.Save()
}

func init() {
	RemoteCmd.AddCommand(setCmd)
	RemoteCmd.AddCommand(renameCmd)
	RemoteCmd.AddCommand(removeCmd)
	RemoteCmd.AddCommand(getURLCmd)
	RemoteCmd.AddCommand(setURLCmd)
}
