// +build darwin

package diagnostics

var macExecutables = []*Executable{
	&Executable{
		Description: "System software overview",
		Command:     "system_profiler SPSoftwareDataType",
	},
	&Executable{
		Description: "Hardware overview",
		Command:     "system_profiler SPHardwareDataType",
	},
	&Executable{
		Description: "Memory",
		Command:     "system_profiler SPMemoryDataType",
	},
	&Executable{
		Description: "Firewall settings",
		Command:     "system_profiler SPFirewallDataType",
	},
}

func init() {
	Executables = append(Executables, macExecutables...)
}
