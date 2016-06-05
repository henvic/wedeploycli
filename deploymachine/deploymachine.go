package deploymachine

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/containers"
	"github.com/launchpad-project/cli/deploy"
	"github.com/launchpad-project/cli/projects"
)

// Machine structure
type Machine struct {
	ProjectID    string
	Flags        *deploy.Flags
	Success      []string
	Errors       *Errors
	SuccessMutex sync.Mutex
	ErrorsMutex  sync.Mutex
	dirMutex     sync.Mutex
	queue        sync.WaitGroup
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

func (de Errors) Error() string {
	var msgs = []string{}

	for _, e := range de.List {
		msgs = append(msgs, fmt.Sprintf("%v: %v", e.ContainerPath, e.Error.Error()))
	}

	return fmt.Sprintf("List of errors (format is container path: error)\n%v",
		strings.Join(msgs, "\n"))
}

// All deploys a list of containers on the given context
func All(list []string, df *deploy.Flags) (success []string, err error) {
	var projectID = config.Stores["project"].Get("id")

	var dm = &Machine{
		ProjectID: projectID,
		Flags:     df,
	}

	created, err := projects.ValidateOrCreate(
		filepath.Join(config.Context.ProjectRoot, "/project.json"))

	if created {
		dm.Success = append(dm.Success, "New project "+projectID+" created")
	}

	if err != nil {
		return success, err
	}

	err = dm.Run(list)

	success = dm.Success

	return success, err
}

// Run deployments
func (m *Machine) Run(list []string) (err error) {
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

func (m *Machine) start(container string) {
	var err = m.mountAndDeploy(container)

	if err != nil {
		m.ErrorsMutex.Lock()
		m.Errors.List = append(m.Errors.List, ContainerError{
			ContainerPath: container,
			Error:         err,
		})
		m.ErrorsMutex.Unlock()
	}

	m.queue.Done()
}

func (m *Machine) deploy(deploy *deploy.Deploy) error {
	m.dirMutex.Lock()
	var err = containers.InstallFromDefinition(m.ProjectID,
		deploy.ContainerPath,
		deploy.Container)
	m.dirMutex.Unlock()
	runtime.Gosched()

	if err != nil {
		return err
	}

	return deploy.HooksAndOnly(m.Flags)
}

func (m *Machine) successFeedback(containerID string) {
	var host = "liferay.io"

	if config.Global.Local {
		host = "liferay.local"
	}

	m.SuccessMutex.Lock()
	m.Success = append(m.Success, fmt.Sprintf(
		"Ready! %v.%v.%v", containerID, m.ProjectID, host))
	m.SuccessMutex.Unlock()
}

func (m *Machine) mountAndDeploy(container string) error {
	var deploy, err = deploy.New(container)

	if err != nil {
		return err
	}

	err = m.deploy(deploy)

	if err == nil {
		m.successFeedback(deploy.Container.ID)
	}

	return err
}
