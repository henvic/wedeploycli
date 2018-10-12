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

	&Executable{
		Description: "Internet connection",
		Command:     "ping -n 3 google.com",
		IgnoreError: true,
	},
}

func init() {
	Executables = append(Executables, windowsExecutables...)
}
