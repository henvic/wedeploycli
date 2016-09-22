package list

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/projects"
)

// Filter parameters for the list command
type Filter struct {
	Project    string
	Containers []string
}

// List containers object
type List struct {
	Detailed              bool
	Filter                Filter
	Projects              []projects.Project
	HandleProjectNotFound func() string
	HandleRequestError    func(error) string
	outStream             io.Writer
	watch                 bool
	retry                 int
	preprint              string
	poolingInterval       time.Duration
}

// New creates a list using the values of a passed Filter
func New(filter Filter) *List {
	var l = &List{
		Filter:    filter,
		outStream: os.Stdout,
	}

	l.HandleProjectNotFound = handleProjectNotFound
	l.HandleRequestError = l.handleRequestError

	return l
}

// NewWatcher creates a list watcher for a given List
func NewWatcher(list *List) *Watcher {
	list.poolingInterval = 1 * time.Second
	return &Watcher{
		List:            list,
		PoolingInterval: 200 * time.Millisecond,
	}
}

// Print containers
func (l *List) Print() {
	var err = l.fetch()
	l.clear()

	if err != nil {
		l.preprint = l.HandleRequestError(err)
		fmt.Fprintf(l.outStream, "%v", l.preprint)
		return
	}

	l.retry = 0
	l.printProjects()

	if len(l.Projects) == 0 {
		var noProjectFound = l.HandleProjectNotFound()

		if l.watch {
			l.preprint = noProjectFound
		} else {
			print(noProjectFound)
		}
	}

	fmt.Fprintf(l.outStream, "%v", l.preprint)
}

func handleProjectNotFound() string {
	return "Project not found.\n"
}

func (l *List) clear() {
	l.preprint = ""
}

func (l *List) printf(format string, a ...interface{}) {
	l.preprint += fmt.Sprintf(format, a...)
}

func (l *List) fetch() error {
	l.resetObjects()

	if l.Filter.Project == "" {
		return l.fetchAllProjects()
	}

	return l.fetchOneProject()
}

func (l *List) handleRequestError(err error) string {
	l.retry++
	if !l.watch {
		fmt.Fprintf(os.Stderr, "%v\n", errorhandling.Handle(err))
		os.Exit(1)
	}

	return fmt.Sprintf(color.Format(color.FgHiRed, "%v #%d\n", errorhandling.Handle(err), l.retry))
}

func (l *List) resetObjects() {
	l.Projects = []projects.Project{}
}

func (l *List) fetchAllProjects() error {
	var ps []projects.Project
	var err error

	if l.watch {
		var ctx, cancel = context.WithTimeout(context.Background(), l.poolingInterval)
		ps, err = projects.List(ctx)
		cancel()
	} else {
		ps, err = projects.List(nil)
	}

	if err != nil {
		return err
	}

	for _, p := range ps {
		l.Projects = append(l.Projects, p)
	}

	return err
}

func (l *List) fetchOneProject() error {
	var p projects.Project
	var err error

	if l.watch {
		var ctx, cancel = context.WithTimeout(context.Background(), l.poolingInterval)
		p, err = projects.Get(ctx, l.Filter.Project)
		cancel()
	} else {
		p, err = projects.Get(nil, l.Filter.Project)
	}

	if err != nil {
		return err
	}

	l.Projects = append(l.Projects, p)
	return err
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
	var word string

	switch {
	// TestLink: custom domain should not be shown for local
	case p.CustomDomain == "" || config.Context.Remote == "":
		word += fmt.Sprintf("%v", getProjectDomain(p.ID))
	case !l.Detailed:
		word += fmt.Sprintf("%v ", p.CustomDomain)
	default:
		word += fmt.Sprintf("%v ", p.CustomDomain)
		word += fmt.Sprintf("(%v)", getProjectDomain(p.ID))
	}

	l.printf(word)
	l.printf(" ")
	l.conditionalPad(word, 55)
	l.printf(getFormattedHealth(p.Health) + "\n")
	l.printContainers(p.ID, p.Containers)
}

func (l *List) printContainers(projectID string, cs containers.Containers) {
	var keys = make([]string, 0)
	for k := range cs {
		if len(l.Filter.Containers) != 0 && !inArray(k, l.Filter.Containers) {
			continue
		}

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
	l.printf(color.Format(getHealthForegroundColor(c.Health), " ● "))
	containerDomain := getContainerDomain(projectID, c.ID)
	l.printf("%v ", containerDomain)
	l.conditionalPad(containerDomain, 52)
	l.printInstances(c.Scale)
	t := getType(c.Type)
	l.printf(color.Format(color.FgHiBlack, "%v ", t))
	l.conditionalPad(t, 23)
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
	StopCondition   (func() bool)
	livew           *uilive.Writer
	End             chan bool
	Teardown        func() string
}

// Start for Watcher
func (w *Watcher) Start() {
	sigs := make(chan os.Signal, 1)
	w.End = make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	w.List.watch = true

	w.livew = uilive.New()
	w.List.outStream = w.livew

	go w.watch()

	go func() {
		<-sigs
		fmt.Fprintln(os.Stdout, "")
		w.End <- true
	}()

	<-w.End

	if w.Teardown != nil {
		fmt.Fprintf(w.List.outStream, "%v", w.Teardown())

		if err := w.livew.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}
}

func (w *Watcher) watch() {
p:
	w.List.Print()

	if err := w.livew.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	if w.StopCondition != nil && w.StopCondition() {
		w.End <- true
		return
	}

	time.Sleep(w.PoolingInterval)
	goto p
}

// Stop for Watcher
func (w *Watcher) Stop() {
	w.End <- true
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

func getProjectDomain(projectID string) string {
	return fmt.Sprintf("%v.wedeploy.me", color.Format(color.Bold, "%v", projectID))
}

func getContainerDomain(projectID, containerID string) string {
	return fmt.Sprintf("%v.%v.wedeploy.me", color.Format(color.Bold, "%v", containerID), projectID)
}

func inArray(key string, haystack []string) bool {
	for _, k := range haystack {
		if key == k {
			return true
		}
	}

	return false
}
