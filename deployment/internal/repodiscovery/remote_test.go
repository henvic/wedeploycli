package repodiscovery

import (
	"testing"
)

type remoteCase struct {
	in     string
	out    string
	hasErr bool
}

var cases = []remoteCase{
	remoteCase{
		in:  "http://github.com/example/hello.git",
		out: "http://github.com/example/hello",
	},
	remoteCase{
		in:  "https://github.com/example/hello.git",
		out: "https://github.com/example/hello",
	},
	remoteCase{
		in:  "https://github.com/example/hello",
		out: "https://github.com/example/hello",
	},
	remoteCase{
		in:  "git@github.com:example/hello.git",
		out: "https://github.com/example/hello",
	},
	remoteCase{
		in:  "git://github.com/example/hello.git",
		out: "https://github.com/example/hello",
	},
	remoteCase{
		in:  "ssh://git@github.com:example/hello.git",
		out: "https://github.com/example/hello",
	},
	remoteCase{
		in:  "git@github.com:example/hello.example.git",
		out: "https://github.com/example/hello.example",
	},
	remoteCase{
		in:  "https://github.com/example/hello.example",
		out: "https://github.com/example/hello.example",
	},
	remoteCase{
		in:  "https://username:password@github.com/example/hello.example",
		out: "https://github.com/example/hello.example",
	},
	remoteCase{
		in:  "git+ssh://git@github.com:example/hello.git",
		out: "https://github.com/example/hello",
	},
	remoteCase{
		in:  "git+https://isaacs@github.com/example/hello.git",
		out: "https://github.com/example/hello",
	},
	remoteCase{
		in:  "git+ssh://git@github.com:example/example.git",
		out: "https://github.com/example/example",
	},
	remoteCase{
		in:  "https://user@bitbucket.org/example/hello.git",
		out: "https://bitbucket.org/example/hello",
	},
	remoteCase{
		in:  "https://user:password@bitbucket.org/example/hello",
		out: "https://bitbucket.org/example/hello",
	},
	remoteCase{
		in:  "https://user@bitbucket.org/example/hello",
		out: "https://bitbucket.org/example/hello",
	},
	remoteCase{
		in:  "git@bitbucket.org:example/hello.git",
		out: "https://bitbucket.org/example/hello",
	},
	remoteCase{
		in:  "https://gitlab.com/gitlab-org/gitlab-ce.git",
		out: "https://gitlab.com/gitlab-org/gitlab-ce",
	},
	remoteCase{
		in:  "git@gitlab.com:gitlab-org/gitlab-ce.git",
		out: "https://gitlab.com/gitlab-org/gitlab-ce",
	},
	remoteCase{
		in:     "git@gitlab.com:%gitlab-org/gitlab-ce.git",
		hasErr: true,
	},
	remoteCase{
		in:     "",
		hasErr: true,
	},
	remoteCase{
		in:     "/path/repo",
		hasErr: true,
	},
}

func TestExtractRepoURL(t *testing.T) {
	for _, c := range cases {
		var got, err = ExtractRepoURL(c.in)

		if got != c.out || (err != nil) == !c.hasErr {
			t.Errorf("Expected ExtractRepoURL(%v) = (%v, has error = %v) got (%v, %v) instead instead", c.in, c.out, c.hasErr, got, err)
		}
	}
}
