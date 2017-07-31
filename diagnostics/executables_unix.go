// +build !windows

package diagnostics

var unixExecutables = []*Executable{
	&Executable{
		Description: "Operating system",
		LogFile:     "system",
		Command:     "uname -a",
	},
	&Executable{
		Description: "CPU info",
		LogFile:     "system",
		Command:     "cat /proc/cpuinfo",
		IgnoreError: true,
	},
	&Executable{
		Description: "Memory info",
		LogFile:     "system",
		Command:     "cat /proc/meminfo",
		IgnoreError: true,
	},
	&Executable{
		Description: "Disk usage",
		LogFile:     "system",
		Command:     "df",
	},
	&Executable{
		Description: "Available memory",
		LogFile:     "system",
		Command:     "free -m",
		IgnoreError: true,
	},
	&Executable{
		Description: "Internet connection",
		LogFile:     "system",
		Command:     "ping -c 3 google.com",
		IgnoreError: true,
	},
}

func init() {
	Executables = append(Executables, unixExecutables...)
}
