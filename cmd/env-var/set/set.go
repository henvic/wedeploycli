package add

import (
	"context"
	"errors"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/env-var/internal/commands"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/services"
)

var (
	file    string
	replace bool
)

// Cmd for setting an environment variable
var Cmd = &cobra.Command{
	Use:     "set",
	Aliases: []string{"add"},
	Short:   "Set an environment variable for a given service",
	Example: `  we env-var set key value
  we env-var set key=value`,
	Args:    checkFileAndArgs,
	PreRunE: preRun,
	RunE:    run,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,

	Requires: cmdflagsfromhost.Requires{
		Auth:    true,
		Project: true,
		Service: true,
	},

	PromptMissingService: true,
}

func init() {
	setupHost.Init(Cmd)
	Cmd.Flags().StringVarP(&file, "file", "F", "",
		"Read environment variables from file")
	Cmd.Flags().BoolVar(&replace, "replace", false,
		"Replace set of environment variables")
}

func checkFileAndArgs(cmd *cobra.Command, args []string) error {
	if file != "" && len(args) != 0 {
		return errors.New("can't merge environment variables from a file and process arguments")
	}

	return nil
}

func preRun(cmd *cobra.Command, args []string) error {
	commands.EnvIsDeprecatedWarning(cmd, args)
	return setupHost.Process(context.Background(), we.Context())
}

func readEnvsFromFile(filepath string) ([]string, error) {
	b, err := ioutil.ReadFile(filepath)

	if err != nil {
		return []string{}, err
	}

	var value = string(b)
	value = strings.Replace(value, "\r\n", "\n", -1) // windows...

	return strings.Split(value, "\n"), nil // notice the line ending instead of space
}

func run(cmd *cobra.Command, args []string) (err error) {
	var c = commands.Command{
		SetupHost:      setupHost,
		ServicesClient: services.New(we.Context()),
	}

	if file != "" {
		args, err = readEnvsFromFile(file)

		if err != nil {
			return err
		}

		c.SkipPrompt = true
	}

	ctx := context.Background()

	if replace {
		return c.Replace(ctx, args)
	}

	return c.Add(ctx, args)
}
