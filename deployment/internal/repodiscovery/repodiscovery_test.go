package repodiscovery

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wedeploy/cli/inspector"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func TestMain(m *testing.M) {
	cleanup()
	setup()

	ec := m.Run()
	cleanup()
	os.Exit(ec)
}

func cleanup() {
	var list = []string{
		"project",
		"project-with-git",
		"service-with-git",
		"service-with-git-init-only",
		"service-without-git",
	}

	for _, mock := range list {
		if err := os.RemoveAll(filepath.Join("mocks", mock)); err != nil {
			panic(err)
		}
	}
}

func setup() {
	cmd := exec.Command("tar", "xjf", "mocks.tar.bz2")
	cmd.Dir = "./mocks"
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

type repoCase struct {
	path         string
	repositories []Repository
	repoless     []string
	err          error
}

var repoCases = []repoCase{
	repoCase{
		path: "mocks/project",
		repositories: []Repository{
			Repository{
				Services:          []string{"service1"},
				Path:              "service-with-git-1",
				Commit:            "b0e71920287fc3d6103f45d173fd0c94cefdcf76",
				CommitAuthor:      "Henrique Vicente",
				CommitAuthorEmail: "henriquevicente@gmail.com",
				CommitMessage:     "Adding wedeploy.json.\n",
				CommitDate:        "2018-08-23 19:16:31 -0400 -0400",
				Branch:            "master",
				CleanWorkingTree:  true,
			},
			Repository{
				Services:          []string{"service2"},
				Path:              "service-with-git-2",
				Commit:            "259a4e4c8bf201113ff828fa83fb7205e7ac7bba",
				CommitAuthor:      "Henrique Vicente",
				CommitAuthorEmail: "henriquevicente@gmail.com",
				CommitMessage:     "Adding service 2.\n",
				CommitDate:        "2018-08-23 19:37:40 -0400 -0400",
				Branch:            "master",
				CleanWorkingTree:  true,
			},
		},
		repoless: []string{"nogit"},
	},
	repoCase{
		path: "mocks/project-with-git",
		repositories: []Repository{
			Repository{
				Services:          []string{"service", "nested"},
				Path:              "",
				Origin:            "https://github.com/example/project-with-git",
				Commit:            "c03668d5c9d781188f5e60e7417d79ccc74b3549",
				CommitAuthor:      "Henrique Vicente",
				CommitAuthorEmail: "henriquevicente@gmail.com",
				CommitMessage:     "Adding project.\n",
				CommitDate:        "2018-08-24 02:19:11 -0400 -0400",
				Branch:            "master",
			},
		},
	},
	repoCase{
		path: "mocks/service-with-git",
		repositories: []Repository{
			Repository{
				Services:          []string{"withgit"},
				Path:              "",
				Origin:            "https://github.com/example/service-with-git",
				Commit:            "ef1a4f3d06b0c3f547f69a3b530c06687efbf4d8",
				CommitAuthor:      "Henrique Vicente",
				CommitAuthorEmail: "henriquevicente@gmail.com",
				CommitMessage:     "Adding service.\n",
				CommitDate:        "2018-08-23 19:18:51 -0400 -0400",
				Branch:            "master",
				CleanWorkingTree:  true,
			},
		},
	},
	repoCase{
		path:     "mocks/service-without-git",
		repoless: []string{"without"},
	},
	repoCase{
		path: "mocks/service-with-git-init-only",
		err:  plumbing.ErrReferenceNotFound,
	},
}

func TestDiscover(t *testing.T) {
	for _, rc := range repoCases {
		i := inspector.ContextOverview{}

		if err := i.Load(rc.path); err != nil {
			panic(err)
		}

		d := Discover{
			Path:     rc.path,
			Services: i.Services,
		}

		repositories, repoless, err := d.Run()

		if !reflect.DeepEqual(repositories, rc.repositories) {
			t.Errorf("Expected repositories to be %+v, got %+v instead", rc.repositories, repositories)
		}

		if !reflect.DeepEqual(repoless, rc.repoless) {
			t.Errorf("Expected repoless to be %v, got %v instead", rc.repoless, repoless)
		}

		if err != rc.err {
			t.Errorf("Expected error to be %v, got %v instead", rc.err, err)
		}
	}
}
