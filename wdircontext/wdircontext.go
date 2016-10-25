package wdircontext

import (
	"errors"

	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
)

var (
	// ErrContextNotFound error message
	ErrContextNotFound = errors.New("Context is not set")
)

// GetProjectID from current working directory project
func GetProjectID() (id string, err error) {
	if config.Context == nil {
		return "", ErrContextNotFound
	}

	var path = config.Context.ProjectRoot
	var project *projects.Project

	project, err = projects.Read(path)

	if err != nil {
		return "", err
	}

	return project.ID, err
}

// GetProjectOrContainerID from current working directory container
func GetProjectOrContainerID() (projectID, containerID string, err error) {
	if config.Context == nil {
		return "", "", ErrContextNotFound
	}

	projectID, err = GetProjectID()

	if err == nil {
		var errc error
		containerID, errc = GetContainerID()

		if errc != containers.ErrContainerNotFound {
			err = errc
		}
	}

	return projectID, containerID, err
}

// GetContainerID from current working directory container
func GetContainerID() (id string, err error) {
	if config.Context == nil {
		return "", ErrContextNotFound
	}

	var path = config.Context.ContainerRoot
	var container *containers.Container

	container, err = containers.Read(path)

	if err != nil {
		return "", err
	}

	return container.ID, err
}
