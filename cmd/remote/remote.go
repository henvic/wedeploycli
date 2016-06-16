package cmdremote

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/verbose"
)

// RemoteCmd runs the WeDeploy structure for development locally
var RemoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Configure WeDeploy remotes",
	Run:   remoteRun,
}

var addCmd = &cobra.Command{
	Use:     "add",
	Short:   "Adds a remote named <name> for the repository at <url>",
	Example: "we remote add hk https://hk.example.com/",
	Run:     addRun,
}

var renameCmd = &cobra.Command{
	Use:     "rename",
	Short:   "Rename the remote named <old> to <new>",
	Example: "we remote rename asia hk",
	Run:     renameRun,
}

var removeCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove the remote named <name>",
	Example: "we remote remove hk",
	Run:     removeRun,
}

var getURLCmd = &cobra.Command{
	Use:   "get-url",
	Short: "Retrieves the URLs for a remote",
	Run:   getURLRun,
}

var setURLCmd = &cobra.Command{
	Use:   "set-url",
	Short: "Changes URLs for the remote",
	Run:   setURLRun,
}

func remoteRun(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		println("This command doesn't take arguments.")
		os.Exit(1)
	}

	var remotes = config.Global.Remotes

	for _, k := range remotes.List() {
		switch verbose.Enabled {
		case true:
			var key, _ = remotes.Get(k)
			fmt.Printf("%s\t%s\n", k, key.URL)
		default:
			fmt.Println(k)
		}
	}
}

func addRun(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		println("This command takes 2 arguments.")
		os.Exit(1)
	}

	var global = config.Global
	var remotes = global.Remotes
	var name = args[0]

	if _, ok := remotes.Get(name); ok {
		println("fatal: remote " + name + " already exists.")
		os.Exit(1)
	}

	remotes.Set(name, args[1])
	global.Save()
}

func renameRun(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		println("This command takes 2 arguments.")
		os.Exit(1)
	}

	var global = config.Global
	var remotes = global.Remotes
	var old = args[0]
	var name = args[1]

	var oldRemote, ok = remotes.Get(old)

	if !ok {
		println("fatal: remote " + old + " doesn't exists.")
		os.Exit(1)
	}

	if _, ok := remotes.Get(name); ok {
		println("fatal: remote " + name + " already exists.")
		os.Exit(1)
	}

	remotes.Set(name, oldRemote.URL, oldRemote.Comment)
	remotes.Del(old)
	global.Save()
}

func removeRun(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		println("This command takes 1 argument.")
		os.Exit(1)
	}

	var global = config.Global
	var remotes = global.Remotes
	var name = args[0]

	if _, ok := remotes.Get(name); !ok {
		println("fatal: remote " + name + " doesn't exists.")
		os.Exit(1)
	}

	remotes.Del(name)
	global.Save()
}

func getURLRun(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		println("This command takes 1 argument.")
		os.Exit(1)
	}

	var remotes = config.Global.Remotes
	var name = args[0]
	var remote, ok = remotes.Get(name)

	if !ok {
		println("fatal: remote " + name + " doesn't exists.")
		os.Exit(1)
	}

	fmt.Println(remote.URL)
}

func setURLRun(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		println("This command takes 2 arguments.")
		os.Exit(1)
	}

	var global = config.Global
	var remotes = global.Remotes
	var name = args[0]
	var uri = args[1]

	if _, ok := remotes.Get(name); !ok {
		println("fatal: remote " + name + " doesn't exists.")
		os.Exit(1)
	}

	remotes.Set(name, uri)
	global.Save()
}

func init() {
	RemoteCmd.AddCommand(addCmd)
	RemoteCmd.AddCommand(renameCmd)
	RemoteCmd.AddCommand(removeCmd)
	RemoteCmd.AddCommand(getURLCmd)
	RemoteCmd.AddCommand(setURLCmd)
}
