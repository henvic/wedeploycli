package git

import (
	"fmt"
	"math/rand"
	"testing"
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
