// +build windows

package diagnostics

var windowsExecutables = []*Executable{
	&Executable{
		appendTo: "system",
		name:     "ver",
		arg:      []string{"SPFirewallDataType"},
	},
	&Executable{
		appendTo: "system",
		name:     "systeminfo",
		arg:      []string{"SPFirewallDataType"},
	},
}

func init() {
	Executables = append(Executables, windowsExecutables...)
}
