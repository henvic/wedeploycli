// +build darwin

package diagnostics

var macExecutables = []*Executable{
	&Executable{
		appendTo: "system",
		name:     "system_profiler",
		arg:      []string{"SPSoftwareDataType"},
	},
	&Executable{
		appendTo: "system",
		name:     "system_profiler",
		arg:      []string{"SPHardwareDataType"},
	},
	&Executable{
		appendTo: "system",
		name:     "system_profiler",
		arg:      []string{"SPMemoryDataType"},
	},
	&Executable{
		appendTo: "system",
		name:     "system_profiler",
		arg:      []string{"SPFirewallDataType"},
	},
}

func init() {
	Executables = append(Executables, macExecutables...)
}
