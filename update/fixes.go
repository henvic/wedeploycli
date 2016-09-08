package update

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/wedeploy/cli/user"
	"github.com/wedeploy/cli/verbose"
)

func autocompleteFix() {
	if runtime.GOOS == "windows" {
		return
	}

	oldAutocomplete := filepath.Join(user.GetHomeDir(), "/.we_autocomplete")

	if err := os.Remove(oldAutocomplete); err != nil {
		verbose.Debug("Can't remove old autocomplete script:", err)
	}
}

var fixes = map[string]func(){
	"1.0.0-alpha-26": autocompleteFix,
	"1.0.0-alpha-27": autocompleteFix,
	"1.0.0-alpha-28": autocompleteFix,
	"1.0.0-alpha-29": autocompleteFix,
}

// ApplyFixes applies fixes functions for updating this tool after updates
func ApplyFixes(pastVersion string) {
	var fix, ok = fixes[pastVersion]

	if !ok {
		verbose.Debug("No update fixes found for past version " + pastVersion)
		return
	}

	println("Applying update fixes for past version " + pastVersion)
	fix()
}
