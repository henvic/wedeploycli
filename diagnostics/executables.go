package diagnostics

import "os"

var liferay = os.Args[0]

// Executables is a list of executables
var Executables = []*Executable{
	&Executable{
		Description: "Installed version",
		Command:     liferay + " version",
	},
	&Executable{
		Command: liferay + " who",
	},
	&Executable{
		Description: "Inspecting working directory context",
		Command:     liferay + " inspect context",
	},
	&Executable{
		Description: "Installed git version",
		Command:     "git version",
	},
}
