// +build !windows

package autocomplete

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/wedeploy/cli/user"
)

var script = `
_we()  {
  COMPREPLY=()
  local cur="${COMP_WORDS[COMP_CWORD]}"
  local opts="$(we autocomplete -- ${COMP_WORDS[@]:1})"
  COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
  return 0
}

complete -F _we we
`

var add = `
autocompleteCommand="[ -f ~/.we_autocomplete ] && source ~/.we_autocomplete"

# Install autocomplete for bash
grep -Fxq "$autocompleteCommand" ~/.bashrc && ec=$? || ec=$?

if [ $ec -ne 0 ] ; then
  echo -e "\n# Adding autocomplete for 'we' \n$autocompleteCommand\n" >> ~/.bashrc
fi

# Install autocomplete for zsh
grep -Fxq "$autocompleteCommand" ~/.zshrc && ec=$? || ec=$?

if [ $ec -ne 0 ] ; then
  echo -e "\n# Adding autocomplete for 'we' \n$autocompleteCommand\n" >> ~/.zshrc
fi
`

var scriptPath string

func init() {
	scriptPath = filepath.Join(user.GetHomeDir(), "/.we_autocomplete")
}

func autoInstall() {
	_, err := os.Stat(scriptPath)

	switch {
	case os.IsNotExist(err):
		install()
	case err != nil:
		fmt.Fprintf(os.Stderr, "Autocomplete autoinstall error: %v\n", err)
	}
}

func install() {
	if err := ioutil.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Saving autocomplete script error: %v\n", err)
	}

	if err := run(add); err != nil {
		fmt.Fprintf(os.Stderr, "Error copying autocomplete scripts: %v\n", err)
	}
}

func run(command string) error {
	process := exec.Command("bash", "-c", command)
	process.Stderr = os.Stderr
	process.Stdout = os.Stdout
	return process.Run()
}
