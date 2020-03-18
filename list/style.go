package list

import (
	"fmt"
	"math/bits"
	"strings"
	"time"

	"github.com/henvic/wedeploycli/color"
	"github.com/henvic/wedeploycli/formatter"
	"github.com/henvic/wedeploycli/projects"
	"github.com/henvic/wedeploycli/services"
)

// Printf list
func (l *List) Printf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(l.w, format, a...)
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

	var header string

	if l.SelectNumber {
		header = "#\t"
	}

	header += strings.Join(servicesHeaders, "\t")

	if bits.OnesCount(uint(l.Details)) != 0 {
		header += "\t"
	}

	if l.Details&Instances != 0 {
		header += "Instances\t"
	}

	if l.Details&CPU != 0 {
		header += "CPU\t"
	}

	if l.Details&Memory != 0 {
		header += "Memory\t"
	}

	if l.Details&CreatedAt != 0 {
		header += "Created at\t"
	}

	if formatter.Human {
		header = strings.Replace(header, "\t", "\t     ", -1)
	}

	l.Printf("%s\n", color.Format(color.FgHiBlack, header))
}

func (l *List) printProjectsOnlyHeaders() {
	var header string

	if l.Filter.HideServices && l.SelectNumber {
		header = "#\t"
	}

	header += strings.Join(projectsHeaders, "\t")

	if bits.OnesCount(uint(l.Details)) != 0 {
		header += "\t"
	}

	if l.Details&CreatedAt != 0 {
		header += "Created at"
	}

	if formatter.Human {
		header = strings.Replace(header, "\t", "\t     ", -1)
	}

	l.Printf("%s\n", color.Format(color.FgHiBlack, header))
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
		l.printProjectOnly(p)
		return
	}

	l.printProjectServices(p)
}

func (l *List) printProjectOnly(p projects.Project) {
	var s = fmt.Sprintf("%s\t%s", p.ProjectID, p.Health)

	if l.Details&CreatedAt != 0 {
		s += fmt.Sprintf("\t%v", p.CreatedAtTime().Format(time.RFC822))
	}

	s += "\n"

	l.Printf(s)
}

func (l *List) printProjectServices(p projects.Project) {
	cs := p.Services

	for _, service := range cs {
		if len(l.Filter.Services) != 0 && !inArray(service.ServiceID, l.Filter.Services) {
			continue
		}

		l.printService(p.ProjectID, service)
	}

	var tabs = len(servicesHeaders) - 1

	for _, d := range details {
		if l.Details&d != 0 {
			tabs++
		}
	}

	if len(cs) == 0 {
		if l.SelectNumber {
			l.Printf("-\t")
		}

		l.Printf("%v    \t%v",
			p.ProjectID,
			color.Format(color.FgYellow, "zero services deployed")+
				strings.Repeat("\t", tabs)+
				"\n")
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

	var ds []string

	if l.Details&Instances != 0 {
		ds = append(ds, fmt.Sprintf("%d", s.Scale))
	}

	if l.Details&CPU != 0 {
		ds = append(ds, fmt.Sprintf("%.6v", s.CPU))
	}

	if l.Details&Memory != 0 {
		ds = append(ds, fmt.Sprintf("%.6v MB", s.Memory))
	}

	if l.Details&CreatedAt != 0 {
		ds = append(ds, s.CreatedAtTime().Format(time.RFC822))
	}

	l.Printf("%s\n", strings.Join(ds, "\t"))
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
