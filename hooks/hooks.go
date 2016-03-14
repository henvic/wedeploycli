package hooks

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/launchpad-project/cli/configstore"
	"github.com/launchpad-project/cli/context"

	"github.com/launchpad-project/cli/config"
	"github.com/mitchellh/mapstructure"
)

// Hooks (after / deploy / main action)
type Hooks struct {
	BeforeBuild  string `json:"before_build" mapstructure:"before_build"`
	Build        string `json:"build"`
	AfterBuild   string `json:"after_build" mapstructure:"after_build"`
	BeforeDeploy string `json:"before_deploy" mapstructure:"before_deploy"`
	Deploy       string `json:"deploy"`
	AfterDeploy  string `json:"after_deploy" mapstructure:"after_deploy"`
}

var (
	// ErrMissingHook is used when the main hook action is missing
	ErrMissingHook = errors.New("Missing hook.")

	outStream io.Writer = os.Stdout
	errStream io.Writer = os.Stderr
)

// Build invokes the build hooks
func Build(ctx *context.Context) error {
	var hooks, err = Get(ctx.Scope)

	if err != nil {
		return err
	}

	if hooks.Build == "" {
		return ErrMissingHook
	}

	if hooks.BeforeBuild != "" {
		RunAndExitOnFailure(hooks.BeforeBuild)
	}

	RunAndExitOnFailure(hooks.Build)

	if hooks.AfterBuild != "" {
		RunAndExitOnFailure(hooks.AfterBuild)
	}

	return err
}

// Deploy invokes the deploy hooks
func Deploy(ctx *context.Context) error {
	var hooks, err = Get(ctx.Scope)

	if err != nil {
		return err
	}

	if hooks.Deploy == "" {
		return ErrMissingHook
	}

	if hooks.BeforeDeploy != "" {
		RunAndExitOnFailure(hooks.BeforeDeploy)
	}

	RunAndExitOnFailure(hooks.Deploy)

	if hooks.AfterDeploy != "" {
		RunAndExitOnFailure(hooks.AfterDeploy)
	}

	return err
}

// Get returns the available hooks
func Get(scope string) (Hooks, error) {
	var s = config.Stores[scope]
	var i, err = s.GetInterface("hooks")
	var hooks Hooks

	if err == nil {
		err = mapstructure.Decode(i, &hooks)
	}

	if err == configstore.ErrConfigKeyNotFound {
		err = ErrMissingHook
	}

	return hooks, err
}

// Run a process synchronously inheriting stderr and stdout
func Run(command string) error {
	var parts = strings.SplitN(command, " ", 2)
	var arguments = []string{}

	if len(parts) == 2 {
		arguments = strings.Split(parts[1], " ")
	}

	process := exec.Command(parts[0], arguments...)
	process.Stderr = errStream
	process.Stdout = outStream
	return process.Run()
}

// RunAndExitOnFailure inheriting stderr and stdout, but kill itself on error
func RunAndExitOnFailure(command string) {
	err := Run(command)

	// edge case not shown on coverage, tested on TestRunAndExitOnFailureFailure
	if err != nil {
		switch err.(type) {
		case *exec.ExitError:
			fmt.Fprintf(errStream, "%v\n", err.(*exec.ExitError))
		default:
			fmt.Fprintf(errStream, "%v\n", err)
		}

		os.Exit(1)
	}
}
