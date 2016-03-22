package cmdcontext

import (
	"errors"

	"github.com/launchpad-project/cli/config"
)

var (
	// ErrNotFound error message
	ErrNotFound = errors.New("ID Not found")

	// ErrInvalidArgumentLength error message
	ErrInvalidArgumentLength = errors.New("Unexpected arguments length")
)

// GetProjectID gets the project ID
func GetProjectID(args []string) (projectID string, err error) {
	switch len(args) {
	case 0:
		return getCtxID("project")
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

func getCtxID(store string) (id string, err error) {
	var configStore = config.Stores[store]

	if configStore == nil {
		return "", ErrNotFound
	}

	id, err = configStore.GetString("id")

	if err != nil {
		return "", ErrNotFound
	}

	return id, nil
}

func getCtxProjectAndContainerID() (projectID, containerID string, err error) {
	projectID, err = getCtxID("project")

	if err != nil {
		return projectID, "", err
	}

	containerID, err = getCtxID("container")

	if err != nil {
		return projectID, "", err
	}

	return projectID, containerID, err
}

func getCtxProjectOrContainerID() (projectID, containerID string, err error) {
	projectID, err = getCtxID("project")

	if err == nil {
		containerID, _ = getCtxID("container")
	}

	return projectID, containerID, err
}
