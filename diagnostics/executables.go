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
	&Executable{
		Description: "Listing running services on local machine",
		Command:     we + " list --remote local",
	},
	&Executable{
		Description: "System docker images",
		LogFile:     "docker_images",
		Command:     "docker images",
	},
	&Executable{
		Description: "Docker services",
		LogFile:     "docker_ps",
		Command:     `docker ps -a --format "table {{.ID}}\t{{.Image}}\t{{.Status}}\t{{.Names}}"`,
	},
	&Executable{
		LogFile: "docker_ps",
		Command: "docker ps",
	},
	&Executable{
		Description: "Docker system-wide information",
		LogFile:     "docker_info",
		Command:     "docker info",
	},
	&Executable{
		Description: "Docker network list",
		LogFile:     "docker_network",
		Command:     "docker network ls",
	},
	&Executable{
		Description: "Inspecting WeDeploy docker network",
		LogFile:     "docker_network",
		Command:     "docker network inspect wedeploy",
	},
}
