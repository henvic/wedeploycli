package list

import (
	"context"

	"github.com/wedeploy/cli/projects"
)

func (l *List) fetch() ([]projects.Project, error) {
	if l.Filter.Project == "" {
		return l.fetchAllProjects()
	}

	return l.fetchOneProject()
}

func (l *List) fetchAllProjects() (ps []projects.Project, err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), l.PoolingInterval)

	projectsClient := projects.New(l.wectx)
	ps, err = projectsClient.List(ctx)
	cancel()

	return ps, err
}

func (l *List) fetchOneProject() (ps []projects.Project, err error) {
	var p projects.Project
	var ctx, cancel = context.WithTimeout(context.Background(), l.PoolingInterval)

	projectsClient := projects.New(l.wectx)
	p, err = projectsClient.Get(ctx, l.Filter.Project)

	// make sure to just add project if no error was received
	if err == nil {
		ps = []projects.Project{p}
	}

	cancel()
	return ps, err
}
