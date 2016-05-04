package run

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
)

func TestGetLaunchpadByDockerHost(t *testing.T) {
	var origDH, origDHExists = os.LookupEnv("DOCKER_HOST")
	os.Setenv("DOCKER_HOST", "tcp://foo.bar:1234")

	var host, err = GetLaunchpadHost()

	if host != "foo.bar" || err != nil {
		t.Errorf("Wanted Launchpad host to be foo.bar. "+
			"Got %v, %v instead", host, err)
	}

	if origDHExists {
		os.Setenv("DOCKER_HOST", origDH)
	} else {
		os.Unsetenv("DOCKER_HOST")
	}
}

func TestGetLaunchpadLocalHost(t *testing.T) {
	var origDH, origDHExists = os.LookupEnv("DOCKER_HOST")
	os.Unsetenv("DOCKER_HOST")

	var host, err = GetLaunchpadHost()

	if host != "localhost" || err != nil {
		t.Errorf("Wanted Launchpad host to be localhost. "+
			"Got %v, %v instead", host, err)
	}

	if origDHExists {
		os.Setenv("DOCKER_HOST", origDH)
	} else {
		os.Unsetenv("DOCKER_HOST")
	}
}

func TestGetLaunchpadHostInvalidParseDockerHost(t *testing.T) {
	var origDH, origDHExists = os.LookupEnv("DOCKER_HOST")
	os.Setenv("DOCKER_HOST", "invalid:10")

	var host, err = GetLaunchpadHost()

	if err == nil {
		t.Errorf("Expected invalid host. "+
			"Got %v, %v instead", host, err)
	}

	var eih = err.(ErrInvalidHost)

	if eih.Error() != "Host is invalid: missing port in address" {
		t.Errorf("Invalid error message: %v", eih.Error())
	}

	if origDHExists {
		os.Setenv("DOCKER_HOST", origDH)
	} else {
		os.Unsetenv("DOCKER_HOST")
	}
}

func TestGetLaunchpadHostInvalidDockerHost(t *testing.T) {
	var origDH, origDHExists = os.LookupEnv("DOCKER_HOST")
	os.Setenv("DOCKER_HOST", "http://[::1]:namedport")

	var host, err = GetLaunchpadHost()

	if err == nil {
		t.Errorf("Expected invalid host. "+
			"Got %v, %v instead", host, err)
	}

	if origDHExists {
		os.Setenv("DOCKER_HOST", origDH)
	} else {
		os.Unsetenv("DOCKER_HOST")
	}
}

type ExistsDependencyProvider struct {
	cmd  string
	find bool
}

var ExistsDependencyCases = []ExistsDependencyProvider{
	{"git", true},
	{fmt.Sprintf("not-found-%d", rand.Int()), false},
}

func TestExistsDependency(t *testing.T) {
	for _, c := range ExistsDependencyCases {
		exists := existsDependency(c.cmd)

		if exists != c.find {
			t.Errorf("existsDependency(%v) should return %v", c.cmd, c.find)
		}
	}
}
