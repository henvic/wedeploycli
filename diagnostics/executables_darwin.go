// +build darwin

package diagnostics

var macExecutables = []*Executable{
	&Executable{
		Description: "System software overview",
		LogFile:     "system",
		Command:     "system_profiler SPSoftwareDataType",
	},
	&Executable{
		Description: "Hardware overview",
		LogFile:     "system",
		Command:     "system_profiler SPHardwareDataType",
	},
	&Executable{
		Description: "Memory",
		LogFile:     "system",
		Command:     "system_profiler SPMemoryDataType",
	},
	&Executable{
		Description: "Firewall settings",
		LogFile:     "system",
		Command:     "system_profiler SPFirewallDataType",
	},
}

func init() {
	Executables = append(Executables, macExecutables...)
}
