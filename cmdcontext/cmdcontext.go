package cmdcontext

import (
	"errors"

	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
)

var (
	// ErrContextNotFound error message
	ErrContextNotFound = errors.New("Context is not set")

	// ErrInvalidArgumentLength error message
	ErrInvalidArgumentLength = errors.New("Unexpected arguments length")
)

// GetProjectID gets the project ID
func GetProjectID(args []string) (projectID string, err error) {
	switch len(args) {
	case 0:
		return getCtxProjectID()
	case 1:
		return args[0], nil
	default:
		return "", ErrInvalidArgumentLength
	}
}

// GetProjectAndContainerID gets the project and container IDs
func GetProjectAndContainerID(args []string) (projectID, containerID string, err error) {
	switch len(args) {
	case 0:
		return getCtxProjectAndContainerID()
	case 2:
		return args[0], args[1], nil
	default:
		return "", "", ErrInvalidArgumentLength
	}
}

// GetProjectOrContainerID gets the project or the project and container IDs
func GetProjectOrContainerID(args []string) (projectID, containerID string, err error) {
	switch len(args) {
	case 0:
		return getCtxProjectOrContainerID()
	case 1:
		return args[0], "", nil
	case 2:
		return args[0], args[1], nil
	default:
		return "", "", ErrInvalidArgumentLength
	}
}

// SplitArguments splits a group of arguments (e.g., project + container)
func SplitArguments(recArgs []string, offset, limit int) []string {
	if len(recArgs) < limit {
		limit = len(recArgs)
	}

	var end = offset + limit

	if end > len(recArgs) {
		end = len(recArgs)
	}

	recArgs = recArgs[offset:end]

	var c = make([]string, end-offset)

	copy(c, recArgs)

	return c
}

func getCtxProjectID() (id string, err error) {
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

func getCtxContainerID() (id string, err error) {
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

func getCtxProjectAndContainerID() (projectID, containerID string, err error) {
	projectID, err = getCtxProjectID()

	if err != nil {
		return projectID, "", err
	}

	containerID, err = getCtxContainerID()

	if err != nil {
		return projectID, "", err
	}

	return projectID, containerID, err
}

func getCtxProjectOrContainerID() (projectID, containerID string, err error) {
	projectID, err = getCtxProjectID()

	if err == nil {
		containerID, _ = getCtxContainerID()
	}

	return projectID, containerID, err
}
