package list

import (
	"fmt"
	"strings"

	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/formatter"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/services"
)

// Printf list
func (l *List) Printf(format string, a ...interface{}) {
	fmt.Fprintf(l.w, format, a...)
}

var projectsHeaders = []string{
	"Project",
	"Status",
}

var servicesHeaders = []string{
	"Project",
	"Service",
	"Image",
	"Status",
}

var detailedServicesHeaders = []string{
	"Instances",
	"CPU",
	"Memory",
}

func (l *List) printServicesHeaders() {
	var projects = l.Projects

	var has bool

	for _, p := range projects {
		if len(p.Services) != 0 {
			has = true
			break
		}
	}

	if !has {
		l.Printf("%s\n", color.Format(color.FgHiBlack, "Project"))
		return
	}

	var servicesHeader string

	if l.SelectNumber {
		servicesHeader = "#\t"
	}

	servicesHeader += strings.Join(servicesHeaders, "\t")

	if l.Detailed {
		servicesHeader += "\t" + strings.Join(detailedServicesHeaders, "\t")
	}

	if formatter.Human {
		servicesHeader = strings.Replace(servicesHeader, "\t", "\t     ", -1)
	}

	l.Printf("%s\n", color.Format(color.FgHiBlack, servicesHeader))
}

func (l *List) printProjectsOnlyHeaders() {
	var projectsHeader string

	if l.Filter.HideServices && l.SelectNumber {
		projectsHeader = "#\t"
	}

	projectsHeader += strings.Join(projectsHeaders, "\t")

	if l.Detailed {
		projectsHeader += "\t" + strings.Join(detailedServicesHeaders, "\t")
	}

	if formatter.Human {
		projectsHeader = strings.Replace(projectsHeader, "\t", "\t     ", -1)
	}

	l.Printf("%s\n", color.Format(color.FgHiBlack, projectsHeader))
}

func (l *List) printProjects() {
	l.selectors = []Selection{}

	l.watchMutex.RLock()
	var projects = l.Projects
	l.watchMutex.RUnlock()

	if len(projects) == 0 {
		l.Printf("No project found.\n")
		return
	}

	switch l.Filter.HideServices {
	case true:
		l.printProjectsOnlyHeaders()
	default:
		l.printServicesHeaders()
	}

	for _, p := range projects {
		l.printProject(p)
	}
}

func (l *List) printProject(p projects.Project) {
	if l.SelectNumber && l.Filter.HideServices {
		l.selectors = append(l.selectors, Selection{
			Project: p.ProjectID,
		})

		l.Printf("%d\t", len(l.selectors))
	}

	if l.Filter.HideServices {
		l.Printf("%v    \t%v\n", p.ProjectID, p.Health)
		return
	}

	cs := p.Services

	for _, service := range cs {
		if len(l.Filter.Services) != 0 && !inArray(service.ServiceID, l.Filter.Services) {
			continue
		}

		l.printService(p.ProjectID, service)
	}

	var tabs = strings.Repeat("\t", len(servicesHeaders)-1)

	if l.Detailed {
		tabs = strings.Repeat("\t", len(detailedServicesHeaders)+1)
	}

	if len(cs) == 0 {
		if l.SelectNumber {
			l.Printf("-\t")
		}

		l.Printf("%v    \t%v",
			p.ProjectID,
			color.Format(color.FgYellow, "zero services deployed")+tabs+"\n")
		return
	}
}

func (l *List) printService(projectID string, s services.Service) {
	if l.SelectNumber {
		l.selectors = append(l.selectors, Selection{
			Project: projectID,
			Service: s.ServiceID,
		})

		l.Printf("%d\t", len(l.selectors))
	}

	l.Printf("%v\t%v\t", projectID, l.getServiceDomain(projectID, s.ServiceID))
	l.Printf("%v\t", s.Type())
	l.Printf("%v\t", s.Health)
	l.printInstances(s.Scale)

	if l.Detailed {
		l.Printf("%.6v\t%.6v MB", s.CPU, s.Memory)
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
	return fmt.Sprintf("%v-%v.%v", serviceID, projectID, l.wectx.ServiceDomain())
}

func inArray(key string, haystack []string) bool {
	for _, k := range haystack {
		if key == k {
			return true
		}
	}

	return false
}
