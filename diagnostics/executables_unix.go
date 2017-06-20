// +build !windows

package diagnostics

var unixExecutables = []*Executable{
	&Executable{
		Description: "Checking operating system",
		LogFile:     "system",
		Command:     "uname -a",
	},
	&Executable{
		Description: "Checking CPU info",
		LogFile:     "system",
		Command:     "cat /proc/cpuinfo",
		IgnoreError: true,
	},
	&Executable{
		Description: "Checking memory info",
		LogFile:     "system",
		Command:     "cat /proc/meminfo",
		IgnoreError: true,
	},
	&Executable{
		Description: "Checking disk usage",
		LogFile:     "system",
		Command:     "df",
	},
	&Executable{
		Description: "Checking available memory",
		LogFile:     "system",
		Command:     "free -m",
		IgnoreError: true,
	},
}

func init() {
	Executables = append(Executables, unixExecutables...)
}
