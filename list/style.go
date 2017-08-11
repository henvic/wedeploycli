package list

import (
	"context"
	"fmt"
	"strings"

	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/fancy"
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
		l.Printf(fmt.Sprintln(fancy.Error("no service found")))
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

func (l *List) printService(projectID string, c services.Service) {
	l.Printf("%v\t%v\t", projectID, l.getServiceDomain(projectID, c.ServiceID))
	l.Printf("%v\t", c.ImageHint)
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

func getHealthForegroundColor(s string) color.Attribute {
	var foregroundMap = map[string]color.Attribute{
		"empty":   color.FgGreen,
		"up":      color.FgHiGreen,
		"warn":    color.FgHiYellow,
		"down":    color.FgHiRed,
		"unknown": color.FgWhite,
	}

	var bg, bgok = foregroundMap[s]

	if !bgok {
		bg = color.FgBlue
	}

	return bg
}

func getHealthBackgroundColor(s string) color.Attribute {
	var backgroundMap = map[string]color.Attribute{
		"empty":   color.BgGreen,
		"up":      color.BgHiGreen,
		"warn":    color.BgHiYellow,
		"down":    color.BgHiRed,
		"unknown": color.BgWhite,
	}

	var bg, bgok = backgroundMap[s]

	if !bgok {
		bg = color.BgHiBlue
	}

	return bg
}

func pad(space int) string {
	return strings.Join(make([]string, space), " ")
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
