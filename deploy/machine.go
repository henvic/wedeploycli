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

	err = installContainerDefinition(m.ProjectID, deploy, m.Flags)

	if err == nil {
		m.SuccessMutex.Lock()
		m.Success = append(m.Success, fmt.Sprintf(
			"Ready! %v.%v.liferay.io", deploy.Container.ID, m.ProjectID))
		m.SuccessMutex.Unlock()
	}

	return err
}

func installContainerDefinition(projectID string, deploy *Deploy, df *Flags) error {
	var err = containers.InstallFromDefinition(projectID, deploy.Container)

	if err != nil {
		return err
	}

	verbose.Debug(deploy.Container.ID, "container definition installed")

	return deploy.HooksAndOnly(df)
}
