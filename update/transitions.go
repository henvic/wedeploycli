package update

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/coreos/go-semver/semver"
	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/user"
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
		return isPastVersionOlder(pastVersion, "1.0.0-alpha-29")
	},
	Apply: func() error {
		if runtime.GOOS == "windows" {
			return nil
		}

		oldAutocomplete := filepath.Join(user.GetHomeDir(), "/.we_autocomplete")

		if err := os.Remove(oldAutocomplete); err != nil {
			return errwrap.Wrapf("Can not remove old autocomplete script: {{err}}", err)
		}

		return nil
	},
}

var transitions = []Transition{
	autocompleteTransition,
}

func isPastVersionOlder(pastVersion, sinceVersion string) bool {
	var old, errOld = semver.NewVersion(sinceVersion)

	if errOld != nil {
		fmt.Fprintf(os.Stderr, "Error parsing \"since\" version: %v\n", errOld)
		return false
	}

	var past, errNew = semver.NewVersion(pastVersion)

	if errNew != nil {
		fmt.Fprintf(os.Stderr, "Error parsing past version %v: %v\n", past, errNew)
		return false
	}

	if past.Compare(*old) > 0 {
		return false
	}

	return true
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
				fmt.Fprintf(os.Stderr, "Error trying to apply transition \"%v\" for past version %v: %v\n",
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
