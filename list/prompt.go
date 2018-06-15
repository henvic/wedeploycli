package list

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/projects"
)

// PromptProject from the list selection
func (l *List) PromptProject(ctx context.Context, wectx config.Context) (*Selection, error) {
	switch l.AllowCreateProjectOnPrompt {
	case true:
		fmt.Printf("Please %s a project from the list below or create a new one.\n",
			color.Format(color.FgMagenta, color.Bold, "select"))
	default:
		fmt.Printf("Please %s a project from the list below.\n",
			color.Format(color.FgMagenta, color.Bold, "select"))
	}

	l.SelectNumber = true
	l.Filter.HideServices = true

	if err := l.Once(ctx, wectx); err != nil {
		return nil, err
	}

	fmt.Println("")
	fmt.Println(fancy.Question("Type project ID or #:"))

	option, err := fancy.Prompt()

	if err != nil {
		return nil, err
	}

	l.watchMutex.RLock()
	var projects = l.Projects
	l.watchMutex.RUnlock()

	for _, p := range projects {
		if option == p.ProjectID {
			return &Selection{
				Project: option,
			}, nil
		}
	}

	selection, err := l.getSelection(option)

	if err != nil && l.AllowCreateProjectOnPrompt {
		return &Selection{
			Project: option,
		}, nil
	}

	if err == nil {
		fmt.Println("")
	}

	return selection, err
}

// PromptProjectOrService from the list selection
func (l *List) PromptProjectOrService(ctx context.Context, wectx config.Context) (*Selection, error) {
	fmt.Printf("Please %s a project or a service from the list below.\n",
		color.Format(color.FgHiMagenta, "select"))
	l.SelectNumber = true

	if err := l.Once(ctx, wectx); err != nil {
		return nil, err
	}

	fmt.Println("")
	fmt.Println(fancy.Question("Type project/service ID or service #:"))

	var option, err = fancy.Prompt()

	if err != nil {
		return nil, err
	}

	if option == "" {
		return nil, errors.New("no selection")
	}

	selection, err := l.selectPromptProjectOrService(option, l.Projects)

	if err == nil {
		fmt.Println("")
	}

	return selection, err
}

func (l *List) selectPromptProjectOrService(option string, projects []projects.Project) (*Selection, error) {
	nextProject, nextService, err := chooseSelectionForProjectOrService(option, projects)

	if err != nil {
		return nil, err
	}

	if nextProject != nil && nextService != nil {
		return dedupPromptProjectOrService(nextProject, nextService)
	}

	if nextProject != nil {
		return nextProject, nil
	}

	if nextService != nil {
		return nextService, nil
	}

	return l.getSelection(option)
}

func chooseSelectionForProjectOrService(option string, projects []projects.Project) (nextProject, nextService *Selection, err error) {
	for _, p := range projects {
		if option == p.ProjectID {
			nextProject = &Selection{
				Project: option,
			}
		}

		for _, s := range p.Services {
			if option == s.ServiceID {
				if nextService != nil {
					return nil, nil, errors.New("multiple services with same ID, use # instead")
				}

				nextService = &Selection{
					Project: p.ProjectID,
					Service: s.ServiceID,
				}
			}
		}
	}

	return nextProject, nextService, nil
}

func dedupPromptProjectOrService(projectCandidate, serviceCandidate *Selection) (*Selection, error) {
	var options = fancy.Options{}
	options.Add("1", "project")
	options.Add("2", "service")

	switch option, err := options.Ask("There is both a service and a project with the same ID."); option {
	case "1", "p", "project":
		return projectCandidate, nil
	case "2", "s", "service":
		return serviceCandidate, nil
	default:
		return nil, err
	}
}

// PromptService from the list selection
func (l *List) PromptService(ctx context.Context, wectx config.Context) (*Selection, error) {
	fmt.Printf("Please %s a service from the list below.\n",
		color.Format(color.FgHiMagenta, "select"))
	l.SelectNumber = true

	if err := l.Once(ctx, wectx); err != nil {
		return nil, err
	}

	fmt.Println("")
	fmt.Println(fancy.Question("Type service ID or #:"))

	var option, err = fancy.Prompt()

	if err != nil {
		return nil, err
	}

	if option == "" {
		return nil, errors.New("no selection")
	}

	selection, err := l.selectPromptService(option, l.Projects)

	if err == nil {
		fmt.Println("")
	}

	return selection, err
}

func (l *List) selectPromptService(option string, projects []projects.Project) (*Selection, error) {
	var nextService *Selection

	for _, p := range l.Projects {
		for _, s := range p.Services {
			if option == s.ServiceID {
				if nextService != nil {
					return nil, errors.New("multiple services with same ID, use # instead")
				}

				nextService = &Selection{
					Project: p.ProjectID,
					Service: s.ServiceID,
				}
			}
		}
	}

	if nextService != nil {
		return nextService, nil
	}

	return l.getSelection(option)
}

func (l *List) getSelection(option string) (*Selection, error) {
	num, err := strconv.Atoi(option)

	if err != nil {
		return nil, errwrap.Wrapf("selection not found", err)
	}

	var sel = l.selectors

	if len(sel) < num {
		return nil, errors.New("invalid selection")
	}

	var s = sel[num-1]
	return &s, nil
}
