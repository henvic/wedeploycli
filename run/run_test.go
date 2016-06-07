package run

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/wedeploy/cli/globalconfigmock"
	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

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

func TestReset(t *testing.T) {
	globalconfigmock.Setup()
	servertest.Setup()

	servertest.Mux.HandleFunc("/reset", tdata.ServerHandler(""))

	var err = Reset()

	if err != nil {
		t.Errorf("Unexpected error %v on reset", err)
	}

	globalconfigmock.Teardown()
	servertest.Teardown()
}
