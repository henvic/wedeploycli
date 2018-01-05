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

var headers = []string{
	"Project",
	"Service",
	"Image",
	"Status",
}

var detailedHeaders = []string{
	"Instances",
	"CPU",
	"Memory",
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

	var header string

	if l.SelectNumber {
		header = "#\t"
	}

	header += strings.Join(headers, "\t")

	if l.Detailed {
		header += "\t" + strings.Join(detailedHeaders, "\t")
	}

	if formatter.Human {
		header = strings.Replace(header, "\t", "\t     ", -1)
	}

	l.Printf("%s\n", color.Format(color.FgHiBlack, header))

	for _, p := range projects {
		l.printProject(p)
	}
}

func (l *List) printProject(p projects.Project) {
	cs := p.Services
	for _, service := range cs {
		if len(l.Filter.Services) != 0 && !inArray(service.ServiceID, l.Filter.Services) {
			continue
		}

		l.printService(p.ProjectID, service)
	}

	var tabs = strings.Repeat("\t", len(headers)-1)

	if l.Detailed {
		tabs = strings.Repeat("\t", len(detailedHeaders)+1)
	}

	if len(cs) == 0 {
		l.Printf("%v    \t%v",
			p.ProjectID,
			color.Format(color.FgYellow, "zero services deployed")+tabs+"\n")
		return
	}
}

func (l *List) printImage(s services.Service) {
	var image = s.ImageHint

	if image == "" {
		image = s.Image
	}

	l.Printf("%v\t", image)
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
	l.printImage(s)
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
