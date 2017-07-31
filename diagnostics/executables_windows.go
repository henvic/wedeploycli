// +build windows

package diagnostics

var windowsExecutables = []*Executable{
	&Executable{
		Description: "system version",
		LogFile:     "system",
		Command:     "ver",
	},
	&Executable{
		Description: "system overview",
		LogFile:     "system",
		Command:     "systeminfo",
	},
}

func init() {
	Executables = append(Executables, windowsExecutables...)
}
