package list

import (
	"context"
	"fmt"
	"strings"

	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/services"
)

func (l *List) printProjects() {
	if len(l.Projects) == 0 {
		l.Printf("No project found.\n")
		return
	}

	for i, p := range l.Projects {
		l.printProject(p)

		if i != len(l.Projects)-1 {
			l.Printf("\n")
		}
	}
}

func (l *List) printProject(p projects.Project) {
	l.Printf(color.Format(getHealthForegroundColor(p.Health), "• "))
	l.Printf("Project: %v\n", color.Format(color.FgBlue, p.ProjectID))

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
		l.Printf(fmt.Sprintln(color.Format(color.FgHiRed, "✖") + " no service found"))
		return
	}
}

func (l *List) printService(projectID string, c services.Service) {
	l.Printf(color.Format(getHealthForegroundColor(c.Health), "• "))
	serviceDomain := l.getServiceDomain(projectID, c.ServiceID)
	l.Printf("%v\t", serviceDomain)
	l.printInstances(c.Scale)
	l.Printf(color.Format(color.FgHiBlack, "%v\t", c.ImageHint))
	l.Printf("%v\n", c.Health)
}

func (l *List) printInstances(instances int) {
	if !l.Detailed {
		return
	}

	var s = fmt.Sprintf("%v instance", instances)

	l.Printf(s)

	if instances == 0 || instances > 1 {
		l.Printf("s")
	}

	l.Printf("\t")
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
	return fmt.Sprintf("%v.%v", color.Format(
		color.Bold, "%v-%v", serviceID, projectID), config.Context.ServiceDomain)
}

func inArray(key string, haystack []string) bool {
	for _, k := range haystack {
		if key == k {
			return true
		}
	}

	return false
}
