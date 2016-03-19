package cmdcontext

import (
	"testing"

	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/configstore"
)

type GetProjectProvider struct {
	Args      []string
	ProjectID string
	Err       error
}

type GetContainerProvider struct {
	Args        []string
	ProjectID   string
	ContainerID string
	Err         error
}

var GetProjectIDCases = []GetProjectProvider{
	{[]string{}, "", ErrNotFound},
	{[]string{"x123", "y454"}, "", ErrInvalidArgumentLength},
	{[]string{"x522"}, "x522", nil},
}

var GetProjectAndContainerIDCases = []GetContainerProvider{
	{[]string{}, "", "", ErrNotFound},
	{[]string{"x433"}, "", "", ErrInvalidArgumentLength},
	{[]string{"x544", "y532", "z752"}, "", "", ErrInvalidArgumentLength},
	{[]string{"x211", "y2224"}, "x211", "y2224", nil},
}

var GetProjectOrContainerIDCases = []GetContainerProvider{
	{[]string{}, "", "", ErrNotFound},
	{[]string{"x007"}, "x007", "", nil},
	{[]string{"x445", "y445"}, "x445", "y445", nil},
	{[]string{"x695", "y151", "z615"}, "", "", ErrInvalidArgumentLength},
}

var GetProjectIDWithProjectStoreCases = []GetProjectProvider{
	{[]string{}, "146274450430843645", nil},
	{[]string{"x454"}, "x454", nil},
	{[]string{"x341", "y53"}, "", ErrInvalidArgumentLength},
	{[]string{"x634"}, "x634", nil},
}

var GetProjectIDWithInvalidProjectStoreCases = []GetProjectProvider{
	{[]string{}, "", ErrNotFound},
	{[]string{"x484"}, "x484", nil},
	{[]string{"x321", "y625"}, "", ErrInvalidArgumentLength},
	{[]string{"x414"}, "x414", nil},
}

var GetProjectAndContainerIDWithProjectStoreCases = []GetContainerProvider{
	{[]string{}, "146274450430843645", "", ErrNotFound},
	{[]string{"146274450430843645", "mycontainer"}, "146274450430843645", "mycontainer", nil},
	{[]string{"146274450430843645", "x4242"}, "146274450430843645", "x4242", nil},
	{[]string{"x555", "y777"}, "x555", "y777", nil},
	{[]string{"x42"}, "", "", ErrInvalidArgumentLength},
}

var GetProjectAndContainerIDWithProjectAndContainerStoreCases = []GetContainerProvider{
	{[]string{}, "146274450430843645", "mycontainer", nil},
	{[]string{"146274450430843645", "mycontainer"}, "146274450430843645", "mycontainer", nil},
	{[]string{"146274450430843645", "x4242"}, "146274450430843645", "x4242", nil},
	{[]string{"x6432", "y535"}, "x6432", "y535", nil},
	{[]string{"x"}, "", "", ErrInvalidArgumentLength},
}

var GetProjectOrContainerIDWithProjectStoreCases = []GetContainerProvider{
	{[]string{}, "146274450430843645", "", nil},
	{[]string{"146274450430843645", "mycontainer"}, "146274450430843645", "mycontainer", nil},
	{[]string{"146274450430843645", "x4242"}, "146274450430843645", "x4242", nil},
	{[]string{"x145"}, "x145", "", nil},
	{[]string{"x555", "y777"}, "x555", "y777", nil},
	{[]string{"146274450430843645", "x414"}, "146274450430843645", "x414", nil},
	{[]string{"x42"}, "x42", "", nil},
}

var GetProjectOrContainerIDWithProjectOrContainerStoreCases = []GetContainerProvider{
	{[]string{}, "146274450430843645", "mycontainer", nil},
	{[]string{"146274450430843645", "mycontainer"}, "146274450430843645", "mycontainer", nil},
	{[]string{"146274450430843645", "x4242"}, "146274450430843645", "x4242", nil},
	{[]string{"x6543"}, "x6543", "", nil},
	{[]string{"146274450430843645", "x414"}, "146274450430843645", "x414", nil},
	{[]string{"146274450430843645", "mycontainer"}, "146274450430843645", "mycontainer", nil},
	{[]string{"x", "y"}, "x", "y", nil},
	{[]string{"x42"}, "x42", "", nil},
}

func TestGetProjectID(t *testing.T) {
	for _, c := range GetProjectIDCases {
		project, err := GetProjectID(c.Args)

		if project != c.ProjectID {
			t.Errorf("Wanted project %v, got %v instead", c.ProjectID, project)
		}

		if err != c.Err {
			t.Errorf("Wanted error %v, got %v instead", c.Err, err)
		}
	}
}

func TestGetProjectAndContainerID(t *testing.T) {
	for _, c := range GetProjectAndContainerIDCases {
		project, container, err := GetProjectAndContainerID(c.Args)

		if project != c.ProjectID {
			t.Errorf("Wanted project %v, got %v instead", c.ProjectID, project)
		}

		if container != c.ContainerID {
			t.Errorf("Wanted container %v, got %v instead", c.ContainerID, container)
		}

		if err != c.Err {
			t.Errorf("Wanted error %v, got %v instead", c.Err, err)
		}
	}
}

func TestGetProjectOrContainerID(t *testing.T) {
	for _, c := range GetProjectOrContainerIDCases {
		project, container, err := GetProjectOrContainerID(c.Args)

		if project != c.ProjectID {
			t.Errorf("Wanted project %v, got %v instead", c.ProjectID, project)
		}

		if container != c.ContainerID {
			t.Errorf("Wanted container %v, got %v instead", c.ContainerID, container)
		}

		if err != c.Err {
			t.Errorf("Wanted error %v, got %v instead", c.Err, err)
		}
	}
}

func TestGetProjectIDWithProjectStore(t *testing.T) {
	var projectStore = configstore.Store{
		Name: "project",
		Path: "./mocks/project/project.json",
	}

	if err := projectStore.Load(); err != nil {
		panic(err)
	}

	config.Stores["project"] = &projectStore

	for _, c := range GetProjectIDWithProjectStoreCases {
		project, err := GetProjectID(c.Args)

		if project != c.ProjectID {
			t.Errorf("Wanted project %v, got %v instead", c.ProjectID, project)
		}

		if err != c.Err {
			t.Errorf("Wanted error %v, got %v instead", c.Err, err)
		}
	}

	config.Stores["project"] = nil
	config.Stores["container"] = nil
}

func TestGetProjectIDWithInvalidProjectStore(t *testing.T) {
	var projectStore = configstore.Store{
		Name: "project",
		Path: "./mocks/invalid-project/project.json",
	}

	if err := projectStore.Load(); err != nil {
		panic(err)
	}

	config.Stores["project"] = &projectStore

	for _, c := range GetProjectIDWithInvalidProjectStoreCases {
		project, err := GetProjectID(c.Args)

		if project != c.ProjectID {
			t.Errorf("Wanted project %v, got %v instead", c.ProjectID, project)
		}

		if err != c.Err {
			t.Errorf("Wanted error %v, got %v instead", c.Err, err)
		}
	}

	config.Stores["project"] = nil
}

func TestGetProjectAndContainerIDWithProjectStore(t *testing.T) {
	var projectStore = configstore.Store{
		Name: "project",
		Path: "./mocks/project/project.json",
	}

	if err := projectStore.Load(); err != nil {
		panic(err)
	}

	config.Stores["project"] = &projectStore

	for _, c := range GetProjectAndContainerIDWithProjectStoreCases {
		project, container, err := GetProjectAndContainerID(c.Args)

		if project != c.ProjectID {
			t.Errorf("Wanted project %v, got %v instead", c.ProjectID, project)
		}

		if container != c.ContainerID {
			t.Errorf("Wanted container %v, got %v instead", c.ContainerID, container)
		}

		if err != c.Err {
			t.Errorf("Wanted error %v, got %v instead", c.Err, err)
		}
	}

	config.Stores["project"] = nil
}

func TestGetProjectAndContainerIDWithProjectAndContainerStore(t *testing.T) {
	var projectStore = configstore.Store{
		Name: "project",
		Path: "./mocks/project/project.json",
	}

	var containerStore = configstore.Store{
		Name: "container",
		Path: "./mocks/project/container/container.json",
	}

	if err := projectStore.Load(); err != nil {
		panic(err)
	}

	if err := containerStore.Load(); err != nil {
		panic(err)
	}

	config.Stores["project"] = &projectStore
	config.Stores["container"] = &containerStore

	for _, c := range GetProjectAndContainerIDWithProjectAndContainerStoreCases {
		project, container, err := GetProjectAndContainerID(c.Args)

		if project != c.ProjectID {
			t.Errorf("Wanted project %v, got %v instead", c.ProjectID, project)
		}

		if container != c.ContainerID {
			t.Errorf("Wanted container %v, got %v instead", c.ContainerID, container)
		}

		if err != c.Err {
			t.Errorf("Wanted error %v, got %v instead", c.Err, err)
		}
	}

	config.Stores["project"] = nil
	config.Stores["container"] = nil
}

func TestGetProjectOrContainerIDWithProjectStore(t *testing.T) {
	var projectStore = configstore.Store{
		Name: "project",
		Path: "./mocks/project/project.json",
	}

	if err := projectStore.Load(); err != nil {
		panic(err)
	}

	config.Stores["project"] = &projectStore

	for _, c := range GetProjectOrContainerIDWithProjectStoreCases {
		project, container, err := GetProjectOrContainerID(c.Args)

		if project != c.ProjectID {
			t.Errorf("Wanted project %v, got %v instead", c.ProjectID, project)
		}

		if container != c.ContainerID {
			t.Errorf("Wanted container %v, got %v instead", c.ContainerID, container)
		}

		if err != c.Err {
			t.Errorf("Wanted error %v, got %v instead", c.Err, err)
		}
	}

	config.Stores["project"] = nil
}

func TestGetProjectOrContainerIDWithProjectOrContainerStore(t *testing.T) {
	var projectStore = configstore.Store{
		Name: "project",
		Path: "./mocks/project/project.json",
	}

	var containerStore = configstore.Store{
		Name: "container",
		Path: "./mocks/project/container/container.json",
	}

	if err := projectStore.Load(); err != nil {
		panic(err)
	}

	if err := containerStore.Load(); err != nil {
		panic(err)
	}

	config.Stores["project"] = &projectStore
	config.Stores["container"] = &containerStore

	for _, c := range GetProjectOrContainerIDWithProjectOrContainerStoreCases {
		project, container, err := GetProjectOrContainerID(c.Args)

		if project != c.ProjectID {
			t.Errorf("Wanted project %v, got %v instead", c.ProjectID, project)
		}

		if container != c.ContainerID {
			t.Errorf("Wanted container %v, got %v instead", c.ContainerID, container)
		}

		if err != c.Err {
			t.Errorf("Wanted error %v, got %v instead", c.Err, err)
		}
	}

	config.Stores["project"] = nil
	config.Stores["container"] = nil
}
