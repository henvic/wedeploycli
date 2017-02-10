// +build !windows

package diagnostics

var unixExecutables = []*Executable{
	&Executable{
		appendTo: "system",
		name:     "uname",
		arg:      []string{"-a"},
	},
	&Executable{
		appendTo: "system",
		name:     "cat",
		arg:      []string{"/proc/cpuinfo"},
	},
	&Executable{
		appendTo: "system",
		name:     "cat",
		arg:      []string{"/proc/meminfo"},
	},
	&Executable{
		appendTo: "system",
		name:     "df",
		arg:      []string{},
	},
	&Executable{
		appendTo: "system",
		name:     "free",
		arg:      []string{"-m"},
	},
}

func init() {
	Executables = append(Executables, unixExecutables...)
}
