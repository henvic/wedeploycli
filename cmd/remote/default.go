package cmdremote

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/prompt"
)

var defaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Set a default (active) remote to use",
	Example: `we remote default
we remote local
we remote wedeploy`,
	PreRunE: cmdargslen.ValidateCmd(0, 1),
	RunE:    setDefaultRun,
}

func getRemoteFromList() (string, error) {
	var global = config.Global
	var keys = global.Remotes.Keys()
	var m = map[string]int{}

	fmt.Println(`Select a remote to use for the next "we" commands:`)

	for v, k := range keys {
		fmt.Printf("%d) %v\n", v+1, k)
		m[k] = v + 1
	}

	var i, err = prompt.SelectOption(len(keys), m)

	if err != nil {
		return "", err
	}

	return keys[i], nil
}

func setDefaultRun(cmd *cobra.Command, args []string) (err error) {
	var name string
	switch len(args) {
	case 0:
		name, err = getRemoteFromList()
		if err != nil {
			return err
		}
	default:
		name = args[0]
	}

	return saveDefaultRemote(name)
}

func saveDefaultRemote(remote string) error {
	var remotes = config.Global.Remotes
	var keys = remotes.Keys()

	for _, k := range keys {
		if remote == k {
			var g = config.Global
			g.DefaultRemote = remote
			return g.Save()
		}
	}

	return fmt.Errorf(`Remote "%v" not found`, remote)
}

func init() {
	RemoteCmd.AddCommand(defaultCmd)
}
