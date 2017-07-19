package wdircontext

import (
	"errors"

	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/services"
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
	var project *projects.ProjectPackage

	project, err = projects.Read(path)

	if err != nil {
		return "", err
	}

	return project.ID, err
}

// GetProjectOrServiceID from current working directory service
func GetProjectOrServiceID() (projectID, serviceID string, err error) {
	if config.Context == nil {
		return "", "", ErrContextNotFound
	}

	projectID, err = GetProjectID()

	if err == nil {
		var errc error
		serviceID, errc = GetServiceID()

		if errc != services.ErrServiceNotFound {
			err = errc
		}
	}

	return projectID, serviceID, err
}

// GetServiceID from current working directory service
func GetServiceID() (id string, err error) {
	if config.Context == nil {
		return "", ErrContextNotFound
	}

	var path = config.Context.ServiceRoot
	var cp *services.ServicePackage

	cp, err = services.Read(path)

	if err != nil {
		return "", err
	}

	return cp.ID, err
}
