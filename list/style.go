package list

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/formatter"
	"github.com/wedeploy/cli/projects"
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
	var word string

	l.Printf(p.ProjectID)
	l.Printf(formatter.CondPad(word, 55))
	l.Printf(getFormattedHealth(p.Health) + "\n")

	var services, err = p.Services(context.Background())

	if err != nil {
		l.Printf("%v\n", errorhandling.Handle(err))
		return
	}

	l.printContainers(p.ProjectID, services)
}

func (l *List) printContainers(projectID string, cs []containers.Container) {
	for _, container := range cs {
		if len(l.Filter.Containers) != 0 && !inArray(container.ServiceID, l.Filter.Containers) {
			continue
		}

		l.printContainer(projectID, container)
	}

	if len(cs) == 0 {
		l.Printf(fmt.Sprintln(color.Format(color.FgHiRed, " (no container found)")))
		return
	}
}

func (l *List) printContainer(projectID string, c containers.Container) {
	l.Printf(color.Format(getHealthForegroundColor(c.Health), " ● "))
	containerDomain := getContainerDomain(projectID, c.ServiceID)
	l.Printf("%v", containerDomain)
	l.Printf(formatter.CondPad(containerDomain, 52))
	l.printInstances(c.Scale)
	t := getType(c.Type)
	l.Printf(color.Format(color.FgHiBlack, "%v", t))
	l.Printf(formatter.CondPad(t, 23))
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

	l.Printf(formatter.CondPad(s, 15))
}

func getType(t string) string {
	var r = regexp.MustCompile(`(.+?)(\:[^:]*$|$)`)
	var matches = r.FindStringSubmatch(t)

	if len(matches) < 2 {
		return ""
	}

	return matches[1]
}

func getHealthForegroundColor(s string) color.Attribute {
	var foregroundMap = map[string]color.Attribute{
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

func getFormattedHealth(s string) string {
	padding := (12 - len(s)) / 2

	if padding < 2 {
		padding = 2
	}

	p := pad(padding)
	return color.Format(color.FgBlack, getHealthBackgroundColor(s), strings.ToUpper(p+s+p))
}

func getContainerDomain(projectID, containerID string) string {
	return fmt.Sprintf("%v.%v", color.Format(
		color.Bold, "%v-%v", containerID, projectID), config.Context.RemoteAddress)
}

func inArray(key string, haystack []string) bool {
	for _, k := range haystack {
		if key == k {
			return true
		}
	}

	return false
}
