// +build !windows

package diagnostics

var unixExecutables = []*Executable{
	&Executable{
		Description: "Operating system",
		Command:     "uname -a",
	},
	&Executable{
		Description: "CPU info",
		Command:     "cat /proc/cpuinfo",
		IgnoreError: true,
	},
	&Executable{
		Description: "Memory info",
		Command:     "cat /proc/meminfo",
		IgnoreError: true,
	},
	&Executable{
		Description: "Disk usage",
		Command:     "df",
	},
	&Executable{
		Description: "Available memory",
		Command:     "free -m",
		IgnoreError: true,
	},
	&Executable{
		Description: "Network",
		Command:     "ifconfig",
		IgnoreError: true,
	},
	&Executable{
		Description: "IP Address and network location",
		Command:     "curl https://ipinfo.io/ -sS",
	},
	&Executable{
		Description: "Internet connection",
		Command:     "ping -c 3 google.com",
		IgnoreError: true,
	},
}

func init() {
	Executables = append(Executables, unixExecutables...)
}
