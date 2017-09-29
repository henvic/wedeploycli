package list

import (
	"context"
	"fmt"
	"strings"

	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/formatter"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/services"
)

func (l *List) printProjects() {
	if len(l.Projects) == 0 {
		l.Printf("No project found.\n")
		return
	}

	var header = "Project\tService\tImage\tStatus"
	if l.Detailed {
		header += "\tInstances\tCPU\tMemory"
	}

	if formatter.Human {
		header = strings.Replace(header, "\t", "\t     ", -1)
	}

	l.Printf("%s\n", color.Format(color.FgHiBlack, header))

	for _, p := range l.Projects {
		l.printProject(p)
	}

}

func (l *List) printProject(p projects.Project) {
	var services, err = p.Services(context.Background())

	if err != nil {
		l.Printf("%v\n", errorhandling.Handle(err))
		return
	}

	l.printServices(p.ProjectID, services)
}

func (l *List) printServices(projectID string, cs []services.Service) {
	for _, service := range cs {
		if len(l.Filter.Services) != 0 && !inArray(service.ServiceID, l.Filter.Services) {
			continue
		}

		l.printService(projectID, service)
	}

	if len(cs) == 0 {
		l.Printf("%v    \t%v\t    \n",
			projectID,
			color.Format(color.FgYellow, "zero services deployed"))
		return
	}
}

func getHealth(health string) string {
	var h = map[string]string{
		"":        "Waiting",
		"up":      "Online",
		"down":    "Offline",
		"warn":    "Warning",
		"unknown": "Unknown",
	}

	if friendly, ok := h[health]; ok {
		return friendly
	}

	return health
}

func (l *List) printImage(s services.Service) {
	var image = s.ImageHint

	if image == "" {
		image = s.Image
	}

	l.Printf("%v\t", image)
}

func (l *List) printService(projectID string, c services.Service) {
	l.Printf("%v\t%v\t", projectID, l.getServiceDomain(projectID, c.ServiceID))
	l.printImage(c)
	l.Printf("%v\t", getHealth(c.Health))
	l.printInstances(c.Scale)

	if l.Detailed {
		l.Printf("%v\t%v MB", c.CPU, c.Memory)
	}

	l.Printf("\n")
}

func (l *List) printInstances(instances int) {
	if !l.Detailed {
		return
	}

	l.Printf("%d\t", instances)
}

func (l *List) getServiceDomain(projectID, serviceID string) string {
	return fmt.Sprintf("%v-%v.%v", serviceID, projectID, config.Context.ServiceDomain)
}

func inArray(key string, haystack []string) bool {
	for _, k := range haystack {
		if key == k {
			return true
		}
	}

	return false
}
