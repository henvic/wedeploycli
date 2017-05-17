package hooks

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/verbose"
)

// Hooks (after / deploy / main action)
type Hooks struct {
	BeforeBuild string `json:"before_build"`
	Build       string `json:"build"`
	AfterBuild  string `json:"after_build"`
	BeforeStart string `json:"before_start"`
	Start       string `json:"start"`
	AfterStart  string `json:"after_start"`
}

// HookError struct
type HookError struct {
	Command string
	Err     error
}

func (he HookError) Error() string {
	return fmt.Sprintf("Command %v failure: %v", he.Command, he.Err.Error())
}

// Build is 'build' hook
const Build = "build"

// Start is 'start' hook
const Start = "start"

var (
	// ErrMissingHook is used when the hook is missing
	ErrMissingHook = errors.New("missing hook")

	outStream io.Writer = os.Stdout
	errStream io.Writer = os.Stderr
)

// Run invokes the hooks for the given hook type on working directory
func (h *Hooks) Run(hookType string, wdir string, notes ...string) error {
	var owd, err = os.Getwd()

	if err != nil {
		return errwrap.Wrapf("Can not get current working dir on hooks run: {{err}}", err)
	}

	if wdir != "" {
		if err = os.Chdir(wdir); err != nil {
			return err
		}
	}

	switch hookType {
	case "build":
		err = h.runBuild(notes...)
	case "start":
		err = h.runStart(notes...)
	default:
		err = ErrMissingHook
	}

	if wdir != "" {
		if ech := os.Chdir(owd); ech != nil {
			fmt.Fprintf(os.Stderr, "Multiple errors: %v\n", err)
			panic(ech)
		}
	}

	return err
}

func (h *Hooks) runBuild(notes ...string) error {
	if h.Build == "" && (h.BeforeBuild != "" || h.AfterBuild != "") {
		fmt.Fprintf(errStream, "No build hook main action\n")
	}

	return runHook([]step{
		step{"before_build", h.BeforeBuild},
		step{"build", h.Build},
		step{"after_build", h.AfterBuild},
	}, notes)
}

func (h *Hooks) runStart(notes ...string) error {
	if h.Start == "" && (h.BeforeStart != "" || h.AfterStart != "") {
		fmt.Fprintf(errStream, "No start hook main action\n")
	}

	return runHook([]step{
		step{"before_start", h.BeforeStart},
		step{"start", h.Start},
		step{"after_start", h.AfterStart},
	}, notes)
}

type step struct {
	name string
	cmd  string
}

func runHook(steps []step, notes []string) (err error) {
	for _, step := range steps {
		if step.cmd == "" {
			continue
		}

		var feedback = "> "

		if len(notes) != 0 {
			feedback += fmt.Sprintf("%v ", notes)
		}

		feedback += step.name + " : " + step.cmd
		fmt.Fprintf(outStream, "%v\n", feedback)

		if err = Run(step.cmd); err != nil {
			return HookError{
				Command: step.cmd,
				Err:     err,
			}
		}
	}

	return nil
}

// Run a process synchronously inheriting stderr and stdout
func Run(command string) error {
	verbose.Debug("> " + command)
	return run(command)
}

func checkShExists() bool {
	_, err := exec.LookPath("sh")
	return err == nil
}
