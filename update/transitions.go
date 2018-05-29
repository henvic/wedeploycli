package update

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/userhome"
	"github.com/wedeploy/cli/verbose"
)

// Transition for update changes
type Transition struct {
	Description func() string
	Test        func(pastVersion string) bool
	Apply       func() error
}

var autocompleteTransition = Transition{
	Description: func() string {
		return "fix broken autocomplete"
	},
	Test: func(pastVersion string) bool {
		return true
	},
	Apply: func() error {
		if runtime.GOOS == "windows" {
			return nil
		}

		oldAutocomplete := filepath.Join(userhome.GetHomeDir(), "/.we_autocomplete")

		if err := os.Remove(oldAutocomplete); err != nil {
			return errwrap.Wrapf("Can not remove old autocomplete script: {{err}}", err)
		}

		return nil
	},
}

var transitions = []Transition{
	autocompleteTransition,
}

// ApplyTransitions applies transition / fixes functions for updating this tool after updates
// It is assumed that this is called only manually, but it is not guaranteed
func ApplyTransitions(pastVersion string) {
	if len(pastVersion) == 0 || pastVersion == defaults.Version || pastVersion == "master" {
		return
	}

	for _, t := range transitions {
		var description = t.Description()

		switch t.Test(pastVersion) {
		case true:
			verbose.Debug(fmt.Sprintf(
				"Applying transition \"%v\" for past version %v",
				description,
				pastVersion))
			if err := t.Apply(); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error trying to apply transition \"%v\" for past version %v: %v\n",
					description,
					pastVersion,
					err)
			}
		default:
			verbose.Debug(fmt.Sprintf(
				"Skipping transition \"%v\" for past version %v",
				description,
				pastVersion))
		}
	}
}
