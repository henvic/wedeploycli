package diagnostics

import "os"

var we = os.Args[0]

// Executables is a list of executables
var Executables = []*Executable{
	&Executable{
		Command: we + " who",
	},
	&Executable{
		Description: "Installed version",
		Command:     we + " version",
	},
	&Executable{
		Command: we + " --verbose",
	},
	&Executable{
		Description: "Inspecting working directory context",
		Command:     we + " inspect context",
	},
}
