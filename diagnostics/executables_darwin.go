// +build darwin

package diagnostics

var macExecutables = []*Executable{
	&Executable{
		Description: "Checking system software overview",
		LogFile:     "system",
		Command:     "system_profiler SPSoftwareDataType",
	},
	&Executable{
		Description: "Checking hardware overview",
		LogFile:     "system",
		Command:     "system_profiler SPHardwareDataType",
	},
	&Executable{
		Description: "Checking memory",
		LogFile:     "system",
		Command:     "system_profiler SPMemoryDataType",
	},
	&Executable{
		Description: "Checking firewall settings",
		LogFile:     "system",
		Command:     "system_profiler SPFirewallDataType",
	},
}

func init() {
	Executables = append(Executables, macExecutables...)
}
