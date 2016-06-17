package link

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
)

// Machine structure
type Machine struct {
	Project      *projects.Project
	ProjectPath  string
	Success      []string
	Errors       *Errors
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

// All links a containers from a list
func All(projectPath string, list []string) (success []string, err error) {
	project, err := projects.Read(projectPath)

	if err != nil {
		return success, err
	}

	var ml = &Machine{
		Project:     project,
		ProjectPath: projectPath,
	}

	created, err := projects.ValidateOrCreate(
		filepath.Join(projectPath, "/project.json"))

	if created {
		ml.Success = append(ml.Success, "New project "+project.ID+" created")
	}

	if err != nil {
		return success, err
	}

	err = ml.run(list)

	success = ml.Success

	return success, err
}

func (m *Machine) run(list []string) (err error) {
	m.Errors = &Errors{
		List: []ContainerError{},
	}

	m.queue.Add(len(list))

	for _, dir := range list {
		go m.start(dir)
	}

	m.queue.Wait()

	if len(m.Errors.List) != 0 {
		err = m.Errors
	}

	return err
}

func (m *Machine) start(dir string) {
	var err = m.mountAndLink(dir)

	if err != nil {
		m.ErrorsMutex.Lock()
		m.Errors.List = append(m.Errors.List, ContainerError{
			ContainerPath: dir,
			Error:         err,
		})
		m.ErrorsMutex.Unlock()
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

func (m *Machine) successFeedback(containerID string) {
	var host = "liferay.local"

	m.SuccessMutex.Lock()
	m.Success = append(m.Success, fmt.Sprintf(
		"Ready! %v.%v.%v", containerID, m.Project.ID, host))
	m.SuccessMutex.Unlock()
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
