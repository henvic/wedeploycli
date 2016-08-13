package cmdcontext

import (
	"os"
	"reflect"
	"testing"

	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
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

type SplitArgumentsProvider struct {
	ReceivedArgs []string
	Offset       int
	Limit        int
	Args         []string
}

var GetProjectIDCases = []GetProjectProvider{
	{[]string{}, "", ErrContextNotFound},
	{[]string{"x123", "y454"}, "", ErrInvalidArgumentLength},
	{[]string{"x522"}, "x522", nil},
}

var GetProjectAndContainerIDCases = []GetContainerProvider{
	{[]string{}, "", "", ErrContextNotFound},
	{[]string{"x433"}, "", "", ErrInvalidArgumentLength},
	{[]string{"x544", "y532", "z752"}, "", "", ErrInvalidArgumentLength},
	{[]string{"x211", "y2224"}, "x211", "y2224", nil},
}

var GetProjectOrContainerIDCases = []GetContainerProvider{
	{[]string{}, "", "", ErrContextNotFound},
	{[]string{"x007"}, "x007", "", nil},
	{[]string{"x445", "y445"}, "x445", "y445", nil},
	{[]string{"x695", "y151", "z615"}, "", "", ErrInvalidArgumentLength},
}

var GetProjectIDWithProjectStoreCases = []GetProjectProvider{
	{[]string{}, "extraction", nil},
	{[]string{"x454"}, "x454", nil},
	{[]string{"x341", "y53"}, "", ErrInvalidArgumentLength},
	{[]string{"x634"}, "x634", nil},
}

var GetProjectIDWithInvalidProjectStoreCases = []GetProjectProvider{
	{[]string{}, "", projects.ErrInvalidProjectID},
	{[]string{"x484"}, "x484", nil},
	{[]string{"x321", "y625"}, "", ErrInvalidArgumentLength},
	{[]string{"x414"}, "x414", nil},
}

var GetProjectAndContainerIDWithProjectStoreCases = []GetContainerProvider{
	{[]string{}, "extraction", "", containers.ErrContainerNotFound},
	{[]string{"extraction", "mycontainer"}, "extraction", "mycontainer", nil},
	{[]string{"extraction", "x4242"}, "extraction", "x4242", nil},
	{[]string{"x555", "y777"}, "x555", "y777", nil},
	{[]string{"x42"}, "", "", ErrInvalidArgumentLength},
}

var GetProjectAndContainerIDWithProjectAndContainerStoreCases = []GetContainerProvider{
	{[]string{}, "extraction", "mycontainer", nil},
	{[]string{"extraction", "mycontainer"}, "extraction", "mycontainer", nil},
	{[]string{"extraction", "x4242"}, "extraction", "x4242", nil},
	{[]string{"x6432", "y535"}, "x6432", "y535", nil},
	{[]string{"x"}, "", "", ErrInvalidArgumentLength},
}

var GetProjectOrContainerIDWithProjectStoreCases = []GetContainerProvider{
	{[]string{}, "extraction", "", nil},
	{[]string{"extraction", "mycontainer"}, "extraction", "mycontainer", nil},
	{[]string{"extraction", "x4242"}, "extraction", "x4242", nil},
	{[]string{"x145"}, "x145", "", nil},
	{[]string{"x555", "y777"}, "x555", "y777", nil},
	{[]string{"extraction", "x414"}, "extraction", "x414", nil},
	{[]string{"x42"}, "x42", "", nil},
}

var GetProjectOrContainerIDWithProjectOrContainerStoreCases = []GetContainerProvider{
	{[]string{}, "extraction", "mycontainer", nil},
	{[]string{"extraction", "mycontainer"}, "extraction", "mycontainer", nil},
	{[]string{"extraction", "x4242"}, "extraction", "x4242", nil},
	{[]string{"x6543"}, "x6543", "", nil},
	{[]string{"extraction", "x414"}, "extraction", "x414", nil},
	{[]string{"extraction", "mycontainer"}, "extraction", "mycontainer", nil},
	{[]string{"x", "y"}, "x", "y", nil},
	{[]string{"x42"}, "x42", "", nil},
}

var SplitArgumentsCases = []SplitArgumentsProvider{
	{[]string{}, 0, 1, []string{}},
	{[]string{"x", "y", "z"}, 0, 0, []string{}},
	{[]string{"x", "y", "z"}, 0, 1, []string{"x"}},
	{[]string{"abc"}, 0, 0, []string{}},
	{[]string{"cde"}, 0, 1, []string{"cde"}},
	{[]string{"fgh"}, 1, 1, []string{}},
	{[]string{"a", "b"}, 0, 2, []string{"a", "b"}},
	{[]string{"a", "b"}, 0, 1, []string{"a"}},
	{[]string{"a", "b"}, 1, 1, []string{"b"}},
	{[]string{"2 3 4", "1 0 1 dog"}, 1, 1, []string{"1 0 1 dog"}},
	{[]string{"2 3 4 x", "1 0 1"}, 1, 2, []string{"1 0 1"}},
	{[]string{"a", "b", "c", "d", "e"}, 2, 0, []string{}},
	{[]string{"a", "b", "c", "d", "e"}, 2, 2, []string{"c", "d"}},
	{[]string{"a", "b", "c", "d", "e"}, 2, 3, []string{"c", "d", "e"}},
	{[]string{"a", "b", "c", "d", "e", "f", "g"}, 3, 3, []string{"d", "e", "f"}},
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
	var workingDir, _ = os.Getwd()
	chdir("./mocks/project/")
	if err := config.Setup(); err != nil {
		panic(err)
	}

	for _, c := range GetProjectIDWithProjectStoreCases {
		project, err := GetProjectID(c.Args)

		if project != c.ProjectID {
			t.Errorf("Wanted project %v, got %v instead", c.ProjectID, project)
		}

		if err != c.Err {
			t.Errorf("Wanted error %v, got %v instead", c.Err, err)
		}
	}

	chdir(workingDir)
	config.Teardown()
}

func TestGetProjectIDWithInvalidProjectStore(t *testing.T) {
	var workingDir, _ = os.Getwd()
	chdir("./mocks/invalid-project/")
	if err := config.Setup(); err != nil {
		panic(err)
	}

	for _, c := range GetProjectIDWithInvalidProjectStoreCases {
		project, err := GetProjectID(c.Args)

		if project != c.ProjectID {
			t.Errorf("Wanted project %v, got %v instead", c.ProjectID, project)
		}

		if err != c.Err {
			t.Errorf("Wanted error %v, got %v instead", c.Err, err)
		}
	}

	chdir(workingDir)
	config.Teardown()
}

func TestGetProjectAndContainerIDWithProjectStore(t *testing.T) {
	var workingDir, _ = os.Getwd()
	chdir("./mocks/project/")
	if err := config.Setup(); err != nil {
		panic(err)
	}

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

	chdir(workingDir)
	config.Teardown()
}

func TestGetProjectAndContainerIDWithProjectAndContainerStore(t *testing.T) {
	var workingDir, _ = os.Getwd()
	chdir("./mocks/project/container")
	if err := config.Setup(); err != nil {
		panic(err)
	}

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

	chdir(workingDir)
	config.Teardown()
}

func TestGetProjectOrContainerIDWithProjectStore(t *testing.T) {
	var workingDir, _ = os.Getwd()
	chdir("./mocks/project")
	if err := config.Setup(); err != nil {
		panic(err)
	}

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

	chdir(workingDir)
	config.Teardown()
}

func TestGetProjectOrContainerIDWithProjectOrContainerStore(t *testing.T) {
	var workingDir, _ = os.Getwd()
	chdir("./mocks/project/container")
	if err := config.Setup(); err != nil {
		panic(err)
	}

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

	chdir(workingDir)
	config.Teardown()
}

func TestSplitArguments(t *testing.T) {
	for _, c := range SplitArgumentsCases {
		var args = SplitArguments(c.ReceivedArgs, c.Offset, c.Limit)

		if !reflect.DeepEqual(args, c.Args) {
			t.Errorf("Wanted %v normalized slice, got %v instead", c.Args, args)
		}
	}
}

func chdir(dir string) {
	if ech := os.Chdir(dir); ech != nil {
		panic(ech)
	}
}
