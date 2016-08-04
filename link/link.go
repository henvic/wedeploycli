package link

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/verbose"
)

// Machine structure
type Machine struct {
	Project     *projects.Project
	Links       []*Link
	ProjectPath string
	Errors      *Errors
	FErrStream  io.Writer
	ErrorsMutex sync.Mutex
	dirMutex    sync.Mutex
	queue       sync.WaitGroup
	Watcher     *list.Watcher
	list        *list.List
}

// Link holds the information of container to be linked
type Link struct {
	Project       *projects.Project
	Container     *containers.Container
	ContainerPath string
}

// ContainerError struct
type ContainerError struct {
	ContainerPath string
	Error         error
}

// Errors list
type Errors struct {
	List []ContainerError
}

var outStream io.Writer = os.Stdout

// New Container link
func New(project *projects.Project, dir string) (*Link, error) {
	var l = &Link{
		ContainerPath: dir,
	}

	l.Project = project

	c, err := containers.Read(l.ContainerPath)
	l.Container = c

	if err != nil {
		return nil, err
	}

	verbose.Debug("Container ID " + c.ID + " for directory " + dir)

	return l, err
}

func (le Errors) Error() string {
	var msgs = []string{}

	for _, e := range le.List {
		msgs = append(msgs, fmt.Sprintf("%v: %v", e.ContainerPath, e.Error.Error()))
	}

	return fmt.Sprintf("List of errors (format is container path: error)\n%v",
		strings.Join(msgs, "\n"))
}

// Setup prepares a project / container for linking
func (m *Machine) Setup(projectPath string, list []string) error {
	m.Errors = &Errors{
		List: []ContainerError{},
	}

	project, err := projects.Read(projectPath)

	if err != nil {
		return err
	}

	m.Project = project
	m.ProjectPath = projectPath

	err = m.createProject()

	if err != nil {
		return err
	}

	for _, dir := range list {
		m.mount(dir)
	}

	return err
}

// Watch changes due to linking
func (m *Machine) Watch() {
	var cs []string

	for _, l := range m.Links {
		cs = append(cs, l.Container.ID)
	}

	m.list = list.New(list.Filter{
		Project:    m.Project.ID,
		Containers: cs,
	})

	m.Watcher = list.NewWatcher(m.list)
	m.Watcher.StopCondition = m.linkedContainersUp
	m.Watcher.Start()
}

func (m *Machine) linkedContainersUp() bool {
	if len(m.list.Projects) == 0 {
		return false
	}

	var projectWatched = m.list.Projects[0]

	for _, link := range m.Links {
		c, ok := projectWatched.Containers[link.Container.ID]

		if !ok || c.Health != "up" {
			return false
		}
	}

	return true
}

func (m *Machine) createProject() error {
	created, err := projects.ValidateOrCreate(
		filepath.Join(m.ProjectPath, "project.json"))

	if created {
		fmt.Fprintf(outStream, "New project %v created.\n", m.Project.ID)
	}

	if err != nil {
		return err
	}

	return m.condAuthProject()
}

func (m *Machine) condAuthProject() error {
	var authFile = filepath.Join(m.ProjectPath, "auth.json")
	var err = projects.SetAuth(m.Project.ID, authFile)

	if os.IsNotExist(err) {
		verbose.Debug("Jumped uploading auth.json for project: does not exist.")
		return nil
	}

	return err
}

// Run links the containers of the list input
func (m *Machine) Run() {
	m.queue.Add(len(m.Links))
	m.linkAll()
	m.queue.Wait()
}

func (m *Machine) linkAll() {
	for _, cl := range m.Links {
		go m.doLink(cl)
	}
}

func (m *Machine) doLink(cl *Link) {
	var err = m.link(cl)

	if err != nil {
		m.logError(cl.ContainerPath, err)
	}

	m.queue.Done()
}

func (m *Machine) link(l *Link) error {
	m.dirMutex.Lock()
	var err = containers.Link(m.Project.ID,
		l.ContainerPath,
		l.Container)
	m.dirMutex.Unlock()
	runtime.Gosched()

	return err
}

func (m *Machine) logError(dir string, err error) {
	m.ErrorsMutex.Lock()
	m.Errors.List = append(m.Errors.List, ContainerError{
		ContainerPath: dir,
		Error:         err,
	})

	if m.FErrStream != nil {
		fmt.Fprintf(m.FErrStream, "%v/ dir error: %v\n", dir, err)
	}

	m.ErrorsMutex.Unlock()
}

func (m *Machine) mountAll() {

}

func (m *Machine) mount(dir string) {
	var l, err = New(m.Project, filepath.Join(m.ProjectPath, dir))

	if err != nil {
		m.logError(dir, err)
		return
	}

	m.Links = append(m.Links, l)
}
