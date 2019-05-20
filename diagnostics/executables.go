package diagnostics

import "os"

var lcp = os.Args[0]

// Executables is a list of executables
var Executables = []*Executable{
	&Executable{
		Description: "Installed version",
		Command:     lcp + " version",
	},
	&Executable{
		Command: lcp + " who",
	},
	&Executable{
		Description: "Inspecting working directory context",
		Command:     lcp + " inspect context",
	},
	&Executable{
		Description: "Installed git version",
		Command:     "git version",
	},
}
