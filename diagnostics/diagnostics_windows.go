// +build windows

package diagnostics

func getRunCommand(command string) (name string, args []string) {
	switch checkBashExists() {
	case true:
		name = "bash"
		args = append([]string{"-c"}, command)
	default:
		name = "cmd"
		args = append([]string{"/c"}, command)
	}

	return name, args
}
