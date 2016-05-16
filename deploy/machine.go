package deploy

import (
	"fmt"
	"path"
	"sync"

	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/containers"
	"github.com/launchpad-project/cli/progress"
	"github.com/launchpad-project/cli/verbose"
)

// Machine structure
type Machine struct {
	ProjectID    string
	Flags        *Flags
	Success      []string
	Errors       *Errors
	SuccessMutex sync.Mutex
	ErrorsMutex  sync.Mutex
	queue        sync.WaitGroup
}

// New Deploy instance
func New(cpath string) (*Deploy, error) {
	var deploy = &Deploy{
		ContainerPath: path.Join(config.Context.ProjectRoot, cpath),
		progress:      progress.New(cpath),
	}

	var c containers.Container
	var err = containers.GetConfig(deploy.ContainerPath, &c)
	deploy.Container = &c

	if err != nil {
		return nil, err
	}

	return deploy, err
}

// Run deployments
func (m *Machine) Run(list []string) (err error) {
	m.Errors = &Errors{
		List: []ContainerError{},
	}

	for _, dir := range list {
		m.installContainerDefinition(dir)
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

func (m *Machine) mountAndDeploy(container string) error {
	var deploy, err = New(container)

	if err != nil {
		return err
	}

	err = deploy.HooksAndOnly(m.Flags)

	if err == nil {
		var host = "liferay.io"

		if config.Stores["global"].Get("local") {
			host = "local"
		}

		m.SuccessMutex.Lock()
		m.Success = append(m.Success, fmt.Sprintf(
			"Ready! %v.%v.%v", deploy.Container.ID, m.ProjectID, host))
		m.SuccessMutex.Unlock()
	}

	return err
}

func (m *Machine) installContainerDefinition(container string) {
	var deploy, err = New(container)

	if err == nil {
		err = containers.InstallFromDefinition(m.ProjectID, deploy.ContainerPath, deploy.Container)
	}

	if err != nil {
		m.ErrorsMutex.Lock()
		m.Errors.List = append(m.Errors.List, ContainerError{
			ContainerPath: container,
			Error:         err,
		})
		m.ErrorsMutex.Unlock()
		return
	}

	verbose.Debug(deploy.Container.ID, "container definition installed")
}
