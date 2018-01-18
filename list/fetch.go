package list

import (
	"context"
	"time"

	"github.com/wedeploy/cli/projects"
)

func (l *List) fetchProjects() ([]projects.Project, error) {
	if l.Filter.Project == "" {
		return l.fetchAllProjects()
	}

	return l.fetchOneProject()
}

func (l *List) fetchAllProjects() (ps []projects.Project, err error) {
	var ctx, cancel = context.WithTimeout(l.ctx, 30*time.Second)
	defer cancel()

	if l.Filter.HideServices {
		return l.projectsClient.List(ctx)
	}

	return l.projectsClient.ListWithServices(ctx)
}

func (l *List) fetchOneProject() (ps []projects.Project, err error) {
	var ctx, cancel = context.WithTimeout(l.ctx, 30*time.Second)
	defer cancel()

	var p projects.Project

	projectsClient := projects.New(l.wectx)

	switch l.Filter.HideServices {
	case true:
		p, err = projectsClient.Get(ctx, l.Filter.Project)
	default:
		p, err = projectsClient.GetWithServices(ctx, l.Filter.Project)
	}

	// make sure to just add project if no error was received
	if err == nil {
		ps = []projects.Project{p}
	}

	return ps, err
}
