// +build !windows

package autocomplete

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/wedeploy/cli/userhome"
)

var script = `if [ -n "$ZSH_VERSION" ]; then
  autoload -U bashcompinit
  bashcompinit
fi

_liferay()  {
  COMPREPLY=()
  local cur="${COMP_WORDS[COMP_CWORD]}"
  local opts="$(liferay autocomplete -- ${COMP_WORDS[@]:1})"
  COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
  return 0
}

complete -F _liferay liferay
`

var add = `
autocompleteCommand="[ -f ~/.liferaycli_autocomplete ] && source ~/.liferaycli_autocomplete"

# Install autocomplete for bash
if [ -w ~/.bashrc ] ; then
  grep -Fxq "$autocompleteCommand" ~/.bashrc && ec=$? || ec=$?

  if [ $ec -ne 0 ] ; then
    echo -e "\n# Adding autocomplete for 'liferay'\n$autocompleteCommand" >> ~/.bashrc
  fi
fi

# Install autocomplete for zsh
if [ -w ~/.zshrc ] ; then
  grep -Fxq "$autocompleteCommand" ~/.zshrc && ec=$? || ec=$?

  if [ $ec -ne 0 ] ; then
    echo -e "\n# Adding autocomplete for 'liferay'\n$autocompleteCommand" >> ~/.zshrc
  fi
fi
`

var scriptPath string

func init() {
	scriptPath = filepath.Join(userhome.GetHomeDir(), "/.liferaycli_autocomplete")
}

func autoInstall() {
	_, err := os.Stat(scriptPath)

	switch {
	case os.IsNotExist(err):
		install()
	case err != nil:
		_, _ = fmt.Fprintf(os.Stderr, "Autocomplete autoinstall error: %v\n", err)
	}
}

func install() {
	if err := ioutil.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Saving autocomplete script error: %v\n", err)
	}

	if err := run(add); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error copying autocomplete scripts: %v\n", err)
	}
}

func run(command string) error {
	process := exec.Command("bash", "-c", command) // #nosec
	process.Stderr = os.Stderr
	process.Stdout = os.Stdout
	return process.Run()
}
