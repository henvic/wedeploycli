package tiny

import (
	"reflect"
	"testing"

	"github.com/henvic/wedeploycli/deployment/internal/repodiscovery"
)

type convertCase struct {
	info Info
	want Tiny
}

var convertCases = []convertCase{
	convertCase{
		Info{
			Deploy: true,
			Repositories: []repodiscovery.Repository{
				repodiscovery.Repository{
					Services:          []string{"service", "nested"},
					Path:              "mocks/project-with-git",
					Origin:            "https://github.com/example/project-with-git",
					Commit:            "5fb64d23dc767fda01416ba0f89375ee88688a74",
					CommitAuthor:      "Example",
					CommitAuthorEmail: "good@example.com",
					CommitMessage:     "deployment: create microformat to list repos used during a deployment.",
					CommitDate:        "2018-08-27 02:23:14 -0600 -0600",
					Branch:            "master",
				},
			},
		},
		Tiny{
			Deploy: true,
			Commit: &commitInfo{
				Repository:  "https://github.com/example/project-with-git",
				SHA:         "5fb64d23dc767fda01416ba0f89375ee88688a74",
				Branch:      "master",
				Message:     "deployment: create microformat to list repos used during a deployment.",
				AuthorName:  "Example",
				AuthorEmail: "good@example.com",
				Date:        "2018-08-27 02:23:14 -0600 -0600",
			},
		},
	},
	convertCase{
		Info{
			Deploy: true,
			Repositories: []repodiscovery.Repository{
				repodiscovery.Repository{
					Services:          []string{"service"},
					Path:              "mocks/service-with-git",
					Origin:            "https://github.com/example/service-with-git",
					Commit:            "5fb64d23dc767fda01416ba0f89375ee88688a74",
					CommitAuthor:      "Example",
					CommitAuthorEmail: "good@example.com",
					CommitMessage:     "deployment: create microformat to list repos used during a deployment.",
					CommitDate:        "2018-08-27 02:23:14 -0600 -0600",
					Branch:            "master",
				},
				repodiscovery.Repository{
					Services:          []string{"foo"},
					Path:              "mocks/another-service",
					Origin:            "https://github.com/example/another-service",
					Commit:            "4b6865baded49855a6b2c4b1160b7460b4a4c9f7",
					CommitAuthor:      "Example",
					CommitAuthorEmail: "good@example.com",
					CommitMessage:     "Adding https://github.com/src-d/go-git.",
					CommitDate:        "2018-08-24 02:23:14 -0600 -0600",
					Branch:            "master",
				},
			},
		},
		Tiny{
			Deploy: true,
		},
	},
	convertCase{
		Info{
			Repositories: []repodiscovery.Repository{
				repodiscovery.Repository{
					Services:          []string{"service", "nested"},
					Path:              "mocks/project-with-git",
					Origin:            "https://github.com/example/service-with-git",
					Commit:            "ecd69078d879f8c53d68c8d7521b1dd04a1b9cd1",
					CommitAuthor:      "Example",
					CommitAuthorEmail: "good@example.com",
					CommitMessage:     "abcd.",
					CommitDate:        "2018-08-24 02:23:14 -0600 -0600",
					Branch:            "master",
				},
			},
			Repoless: []string{"foo"},
		},
		Tiny{},
	},
}

func TestConvert(t *testing.T) {
	for _, c := range convertCases {
		var got = Convert(c.info)

		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("Expected tiny info to be %+v, got %+v instead", c.want, got)
		}
	}
}
