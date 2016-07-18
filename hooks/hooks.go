package hooks

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Hooks (after / deploy / main action)
type Hooks struct {
	BeforeBuild  string `json:"before_build"`
	Build        string `json:"build"`
	AfterBuild   string `json:"after_build"`
	BeforeDeploy string `json:"before_deploy"`
	Deploy       string `json:"deploy"`
	AfterDeploy  string `json:"after_deploy"`
}

// Build is 'build' hook
const Build = "build"

var (
	// ErrMissingHook is used when the hook is missing
	ErrMissingHook = errors.New("Missing hook.")

	outStream io.Writer = os.Stdout
	errStream io.Writer = os.Stderr
)

// Run invokes the hooks for the given hook type
func (h *Hooks) Run(hookType string) error {
	switch hookType {
	case "build":
		return h.runBuild()
	default:
		return ErrMissingHook
	}
}

func (h *Hooks) runBuild() error {
	if h.Build == "" && (h.BeforeBuild != "" || h.AfterBuild != "") {
		fmt.Fprintf(errStream, "Error: no build hook main action\n")
	}

	if h.BeforeBuild != "" {
		RunAndExitOnFailure(h.BeforeBuild)
	}

	if h.Build != "" {
		RunAndExitOnFailure(h.Build)
	}

	if h.AfterBuild != "" {
		RunAndExitOnFailure(h.AfterBuild)
	}

	return nil
}

// Run a process synchronously inheriting stderr and stdout
func Run(command string) error {
	return run(command)
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
