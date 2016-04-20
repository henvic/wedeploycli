package deploy

import (
	"fmt"
	"os"
	"testing"
)

func TestErrors(t *testing.T) {
	var e error = &Errors{
		List: map[string]error{
			"foo": os.ErrExist,
			"bar": os.ErrNotExist,
		},
	}

	var want = "Deploy error: map[foo:file already exists bar:file does not exist]"

	if fmt.Sprintf("%v", e) != want {
		t.Errorf("Wanted error %v, got %v instead.", e, want)
	}
}
