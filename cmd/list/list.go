package cmdlist

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
)

// ListCmd is used for getting a list of projects and containers
var ListCmd = &cobra.Command{
	Use:   "list or list [project] to filter by project",
	Short: "List projects and containers running on WeDeploy",
	Run:   listRun,
}

type list struct {
	totalContainers      int
	projects             []projects.Project
	containersProjectMap map[string]containers.Containers
	preprint             string
}

func (l *list) printf(format string, a ...interface{}) {
	l.preprint += fmt.Sprintf(format, a...)
}

func (l *list) flush() {
	fmt.Println(l.preprint)
	l.preprint = ""

}

func (l *list) mapContainers() {
	for _, p := range l.projects {
		var cs, err = containers.List(p.ID)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving containers for %v.\n", p.ID)
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		l.containersProjectMap[p.ID] = cs
		l.totalContainers += len(cs)
	}
}

func (l *list) printProjects() {
	for _, p := range l.projects {
		l.printProject(p)
	}

	l.printf("total %v\n", l.totalContainers)
}

func (l *list) printProject(p projects.Project) {
	if p.CustomDomain != "" {
		l.printf("%v (%v)", p.CustomDomain, getProjectDomain(p.ID))
	} else {
		l.printf("%v", getProjectDomain(p.ID))
	}

	l.printf("\t\t" + getFormattedHealth(p.Health) + "\n")
	l.printContainers(p.ID)
}

func (l *list) printContainers(projectID string) {
	var cs = l.containersProjectMap[projectID]
	var keys = make([]string, 0)
	for k := range cs {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		c := cs[k]
		l.printf(color.Format(
			color.FgBlack,
			getHealthForegroundColor(c.Health), "●")+" %v\t",
			getContainerDomain(projectID, c.ID))
		l.printf("%v instance", c.Instances)

		if c.Instances > 1 {
			l.printf("s")
		}

		l.printf(" [%v]", c.Type)
		l.printf(" [%v]\n", c.Health)
	}
}

func getHealthForegroundColor(s string) color.Attribute {
	var foregroundMap = map[string]color.Attribute{
		"up":      color.FgHiGreen,
		"warn":    color.FgHiYellow,
		"down":    color.FgHiBlack,
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
		"down":    color.BgHiBlack,
		"unknown": color.BgWhite,
	}

	var bg, bgok = backgroundMap[s]

	if !bgok {
		bg = color.BgHiBlue
	}

	return bg
}

func getFormattedHealth(s string) string {
	padding := (12 - len(s)) / 2

	if padding < 2 {
		padding = 2
	}

	p := strings.Join(make([]string, padding), " ")
	return color.Format(color.FgBlack, getHealthBackgroundColor(s), strings.ToUpper(p+s+p))
}

func getProjectDomain(projectID string) string {
	return fmt.Sprintf("%v.wedeploy.me", projectID)
}

func getContainerDomain(projectID, containerID string) string {
	return fmt.Sprintf("%v.%v.wedeploy.me", containerID, projectID)
}

func listRun(cmd *cobra.Command, args []string) {
	var l = &list{
		projects:             []projects.Project{},
		containersProjectMap: map[string]containers.Containers{},
	}

	switch len(args) {
	case 0:
		var ps, err = projects.List()

		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		for _, project := range ps {
			l.projects = append(l.projects, project)
		}
	case 1:
		var p, err = projects.Get(args[0])

		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		l.projects = append(l.projects, p)
	default:
		println("This command takes 0 or 1 argument.")
		os.Exit(1)
	}

	l.mapContainers()
	l.printProjects()
	l.flush()
}
