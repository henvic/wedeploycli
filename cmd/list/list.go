package cmdlist

import (
	"fmt"
	"os"
	"regexp"
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
	projects             []projects.Project
	containersProjectMap map[string]containers.Containers
	preprint             string
}

var detailed bool

func (l *list) printf(format string, a ...interface{}) {
	l.preprint += fmt.Sprintf(format, a...)
}

func (l *list) flush() {
	fmt.Print(l.preprint)
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
	}
}

func (l *list) printProjects() {
	for i, p := range l.projects {
		l.printProject(p)

		if i != len(l.projects)-1 {
			l.printf("\n")
		}
	}
}

func (l *list) printProject(p projects.Project) {
	word := fmt.Sprintf("Project " + color.Format(color.Bold, "%v ", p.Name))

	switch {
	case p.CustomDomain == "":
		word += fmt.Sprintf("%v", getProjectDomain(p.ID))
	case !detailed:
		word += fmt.Sprintf("%v ", p.CustomDomain)
	default:
		word += fmt.Sprintf("%v ", p.CustomDomain)
		word += fmt.Sprintf("(%v)", getProjectDomain(p.ID))
	}

	l.printf(word)
	l.printf(" ")
	l.conditionalPad(word, 72)
	l.printf(getFormattedHealth(p.Health) + "\n")
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
		l.printContainer(projectID, cs[k])
	}
}

func (l *list) conditionalPad(word string, maxWord int) {
	if maxWord > len(word) {
		l.printf(pad(maxWord - len(word)))
	}
}

func (l *list) printContainer(projectID string, c *containers.Container) {
	l.printf(color.Format(color.FgBlack, getHealthForegroundColor(c.Health), "● "))
	l.printf("%v ", c.Name)
	l.conditionalPad(c.Name, 20)
	containerDomain := getContainerDomain(projectID, c.ID)
	l.printf("%v ", containerDomain)
	l.conditionalPad(containerDomain, 42)
	l.printInstances(c.Instances)
	t := getType(c.Type)
	l.printf(color.Format(color.FgHiBlack, "%v ", t))
	l.conditionalPad(t, 20)
	l.printf("%v\n", c.Health)
}

func (l *list) printInstances(instances int) {
	if !detailed {
		return
	}

	var s = fmt.Sprintf("%v instance", instances)

	l.printf(s)

	if instances == 0 || instances > 1 {
		l.printf("s")
	}

	l.printf(" ")
	l.conditionalPad(s, 14)
}

func getType(t string) string {
	var r, err = regexp.Compile(`(.+?)(\:[^:]*$|$)`)

	if err != nil {
		panic(err)
	}

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

func getProjectDomain(projectID string) string {
	return fmt.Sprintf("%v.wedeploy.me", color.Format(color.Bold, "%v", projectID))
}

func getContainerDomain(projectID, containerID string) string {
	return fmt.Sprintf("%v.%v.wedeploy.me", color.Format(color.Bold, "%v", containerID), projectID)
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

func init() {
	ListCmd.Flags().BoolVarP(
		&detailed,
		"detailed", "d", false, "Show more containers details.")
}
