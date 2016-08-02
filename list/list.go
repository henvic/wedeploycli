package list

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
)

type List struct {
	Detailed             bool
	Projects             []projects.Project
	containersProjectMap map[string]containers.Containers
	preprint             string
}

func (l *List) Print() {
	l.mapContainers()
	l.printProjects()
	l.flush()
}

func (l *List) printf(format string, a ...interface{}) {
	l.preprint += fmt.Sprintf(format, a...)
}

func (l *List) flush() {
	fmt.Print(l.preprint)
	l.preprint = ""
}

func (l *List) flushString() string {
	var t = l.preprint
	l.preprint = ""
	return t
}

func (l *List) mapContainers() {
	l.containersProjectMap = map[string]containers.Containers{}

	for _, p := range l.Projects {
		var cs, err = containers.List(p.ID)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving containers for %v.\n", p.ID)
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		l.containersProjectMap[p.ID] = cs
	}
}

func (l *List) printProjects() {
	for i, p := range l.Projects {
		l.printProject(p)

		if i != len(l.Projects)-1 {
			l.printf("\n")
		}
	}
}

func (l *List) printProject(p projects.Project) {
	word := fmt.Sprintf("Project " + color.Format(color.Bold, "%v ", p.Name))

	switch {
	case p.CustomDomain == "":
		word += fmt.Sprintf("%v", getProjectDomain(p.ID))
	case !l.Detailed:
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

func (l *List) printContainers(projectID string) {
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

func (l *List) conditionalPad(word string, maxWord int) {
	if maxWord > len(word) {
		l.printf(pad(maxWord - len(word)))
	}
}

func (l *List) printContainer(projectID string, c *containers.Container) {
	l.printf(color.Format(getHealthForegroundColor(c.Health), "● "))
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

func (l *List) printInstances(instances int) {
	if !l.Detailed {
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

// Watcher structure
type Watcher struct {
	List            *List
	PoolingInterval time.Duration
	ticker          *time.Ticker
	livew           *uilive.Writer
}

// Watch logs
func Watch(watcher *Watcher) {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	watcher.livew = uilive.New()
	watcher.Start()

	go func() {
		<-sigs
		fmt.Fprintln(os.Stdout, "")
		watcher.Stop()
		done <- true
	}()

	<-done
}

// Start for Watcher
func (w *Watcher) Start() {
	w.ticker = time.NewTicker(w.PoolingInterval)
	w.livew.Start()

	go func() {
		w.pool()
		for range w.ticker.C {
			w.pool()
		}
	}()
}

// Stop for Watcher
func (w *Watcher) Stop() {
	w.ticker.Stop()
	w.livew.Stop()
	os.Exit(0)
}

func (w *Watcher) pool() {
	w.List.mapContainers()
	w.List.printProjects()
	fmt.Fprintf(w.livew, "%v", w.List.flushString())
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
