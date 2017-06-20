// +build windows

package diagnostics

var windowsExecutables = []*Executable{
	&Executable{
		Description: "Checking system version",
		LogFile:     "system",
		Command:     "ver",
	},
	&Executable{
		Description: "Checking system overview",
		LogFile:     "system",
		Command:     "systeminfo",
	},
}

func init() {
	Executables = append(Executables, windowsExecutables...)
}
