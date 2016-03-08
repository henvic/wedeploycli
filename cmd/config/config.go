package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/configstore"
	"github.com/launchpad-project/cli/prompt"
	"github.com/spf13/cobra"
)

// ConfigCmd is used for configuring the CLI tool and app
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration for the Launchpad CLI tool",
	Run:   configRun,
}

var (
	listArg     bool
	setArg      bool
	confArg     string
	configStore *configstore.Store
)

func listKeys() {
	for key, configurable := range configStore.ConfigurableKeys {
		if configurable {
			fmt.Println(key, "=", configStore.GetRequiredString(key))
		}
	}
}

func configRun(cmd *cobra.Command, args []string) {
	switch confArg {
	case "default":
		configStore = config.Stores[config.Context.Scope]
	case "project", "container", "global":
		configStore = config.Stores[confArg]
	default:
		println("Invalid config scope")
		os.Exit(1)
	}

	if configStore == nil {
		println("Not in the given config scope")
		os.Exit(1)
	}

	if listArg {
		listKeys()
		return
	}

	if len(args) == 0 {
		cmd.Help()
		return
	}

	var key = args[0]
	var value string

	if len(args) != 1 {
		value = strings.Join(args[1:], " ")
	} else if setArg {
		value = prompt.Prompt(key)
	}

	if len(args) != 1 || setArg {
		if err := configStore.SetAndSaveEditableKey(key, value); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		return
	}

	value, err := configStore.GetString(key)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	fmt.Println(value)
}

func init() {
	ConfigCmd.Flags().BoolVarP(&listArg, "list", "l", false, "list all")
	ConfigCmd.Flags().BoolVar(&setArg, "set", false, "set property")
	ConfigCmd.Flags().StringVar(&confArg,
		"conf",
		"default",
		"global, project, container, default (local scope)")
}
