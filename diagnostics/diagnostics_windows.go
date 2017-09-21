// +build windows

package diagnostics

func getRunCommand(command string) (name string, args []string) {
	name = "cmd"
	args = append([]string{"/c"}, command)

	return name, args
}
