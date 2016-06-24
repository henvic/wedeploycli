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
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/verbose"
)

// Machine structure
type Machine struct {
	Project      *projects.Project
	ProjectPath  string
	Success      []string
	Errors       *Errors
	FErrStream   io.Writer
	FOutStream   io.Writer
	SuccessMutex sync.Mutex
	ErrorsMutex  sync.Mutex
	dirMutex     sync.Mutex
	queue        sync.WaitGroup
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
func (m *Machine) Setup(projectPath string) error {
	project, err := projects.Read(projectPath)

	if err != nil {
		return err
	}

	m.Project = project
	m.ProjectPath = projectPath

	return m.createProject()
}

func (m *Machine) createProject() error {
	created, err := projects.ValidateOrCreate(
		filepath.Join(m.ProjectPath, "project.json"))

	if created {
		m.logSuccess("New project " + m.Project.ID + " created")
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
		verbose.Debug("Jumped uploading auth.json for project: does not exist")
		return nil
	}

	return err
}

// Run links the containers of the list input
func (m *Machine) Run(list []string) {
	m.Errors = &Errors{
		List: []ContainerError{},
	}

	m.queue.Add(len(list))

	for _, dir := range list {
		go m.start(dir)
	}

	m.queue.Wait()
}

func (m *Machine) start(dir string) {
	var err = m.mountAndLink(dir)

	if err != nil {
		m.logError(dir, err)
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

func (m *Machine) logSuccess(msg string) {
	m.SuccessMutex.Lock()
	m.Success = append(m.Success, msg)

	if m.FOutStream != nil {
		fmt.Fprintf(m.FOutStream, "%v\n", msg)
	}

	m.SuccessMutex.Unlock()
}

func (m *Machine) successFeedback(containerID string) {
	var host = "wedeploy.local"

	m.logSuccess(fmt.Sprintf("Ready! %v.%v.%v",
		containerID,
		m.Project.ID,
		host))
}

func (m *Machine) mountAndLink(dir string) error {
	var l, err = New(m.Project, filepath.Join(m.ProjectPath, dir))

	if err != nil {
		return err
	}

	err = m.link(l)

	if err == nil {
		m.successFeedback(l.Container.ID)
	}

	return err
}
