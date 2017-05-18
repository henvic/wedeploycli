// +build !nocompile

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
)

func init() {
	var workingDir, err = os.Getwd()

	if err != nil {
		panic(err)
	}

	chdir("..")
	defer chdir(workingDir)
	compile()
}

func compile() {
	build()

	var err error
	binary, err = filepath.Abs("cli")

	if err != nil {
		panic(err)
	}
}

func build() {
	var cmd = exec.Command("go", "build", "-race")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
