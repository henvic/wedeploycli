package remote

import (
	"fmt"

	"github.com/henvic/wedeploycli/color"
	"github.com/henvic/wedeploycli/command/internal/we"
	"github.com/henvic/wedeploycli/prompt"
	"github.com/spf13/cobra"
)

var defaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Set a default (active) remote to use",
	Example: `lcp remote default
lcp remote local
lcp remote wedeploy`,
	Args: cobra.MaximumNArgs(1),
	RunE: setDefaultRun,
}

func getRemoteFromList() (string, error) {
	var wectx = we.Context()
	var conf = wectx.Config()
	var params = conf.GetParams()
	var rl = params.Remotes
	var keys = rl.Keys()

	var m = map[string]int{}

	fmt.Println(`Select a remote to use for the next "lcp" commands:`)
	fmt.Println(color.Format(color.FgHiBlack, "#\tRemote"))

	for v, k := range keys {
		fmt.Printf("%d\t%v\n", v+1, k)
		m[k] = v + 1
	}

	fmt.Print("Choice: ")
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
	var wectx = we.Context()
	var conf = wectx.Config()
	var params = conf.GetParams()
	var remotes = params.Remotes
	var keys = remotes.Keys()

	for _, k := range keys {
		if remote == k {
			params.DefaultRemote = remote
			conf.SetParams(params)
			return conf.Save()
		}
	}

	return fmt.Errorf(`remote "%v" not found`, remote)
}

func init() {
	RemoteCmd.AddCommand(defaultCmd)
}
