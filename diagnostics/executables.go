package diagnostics

import "os"

var we = os.Args[0]

// Executables is a list of executables
var Executables = []*Executable{
	&Executable{
		name: we,
		arg:  []string{"who"},
	},
	&Executable{
		name: we,
		arg:  []string{"version"},
	},
	&Executable{
		name: we,
		arg:  []string{"remote", "-v"},
	},
	&Executable{
		name: we,
		arg:  []string{"inspect", "context"},
	},
	&Executable{
		appendTo: "docker_images",
		name:     "docker",
		arg:      []string{"images"},
	},
	&Executable{
		appendTo: "docker_ps",
		name:     "docker",
		arg:      []string{"ps"},
	},
	&Executable{
		appendTo: "docker_info",
		name:     "docker",
		arg:      []string{"info"},
	},
	&Executable{
		appendTo: "docker_network",
		name:     "docker",
		arg:      []string{"network", "ls"},
	},
	&Executable{
		appendTo: "docker_network",
		name:     "docker",
		arg:      []string{"network", "inspect", "bridge"},
	},
}
