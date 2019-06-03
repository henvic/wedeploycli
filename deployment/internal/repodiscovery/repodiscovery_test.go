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
		"service-with-git-no-upstream",
		"service-without-git",
	}

	for _, mock := range list {
		if err := os.RemoveAll(filepath.Join("mocks", mock)); err != nil {
			panic(err)
		}
	}
}

func setup() {
	if !existsDependency("tar") {
		return
	}

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
				Commit:            "53f46851d481468d34eef9c868eceaad8778e60c",
				CommitAuthor:      "Henrique Vicente",
				CommitAuthorEmail: "henriquevicente@gmail.com",
				CommitMessage:     "Adding LCP.json.\n",
				CommitDate:        "2018-08-23 19:16:31 -0400 -0400",
				Branch:            "master",
				CleanWorkingTree:  true,
			},
			Repository{
				Services:          []string{"service2"},
				Path:              "service-with-git-2",
				Commit:            "c17f3c056d0facfb7d4c04f32d6191deb9b9f95a",
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
				Commit:            "0bf3c655b40eb018978a31f208376c94b2529a08",
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
				Commit:            "e9bfabfb987a4594321a2b7045bd9abae21d69b3",
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
		path: "mocks/service-with-git-no-upstream",
		repositories: []Repository{
			Repository{
				Services:          []string{"withgit"},
				Path:              "",
				Origin:            "https://github.com/example/service-with-git",
				Commit:            "e9bfabfb987a4594321a2b7045bd9abae21d69b3",
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
	if !existsDependency("tar") {
		t.Skip("tar not found in system")
	}

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

func existsDependency(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
