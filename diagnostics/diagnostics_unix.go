// +build !windows

package diagnostics

func getRunCommand(command string) (name string, args []string) {
	name = "bash"
	args = append([]string{"-c"}, command)
	if !checkBashExists() {
		name = "sh"
	}

	return name, args
}
