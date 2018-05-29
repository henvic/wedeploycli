package main

import (
	"fmt"
	"os"

	"github.com/wedeploy/cli/update/releasenotes"
)

func main() {
	if len(os.Args) != 2 {
		_, _ = fmt.Fprint(os.Stderr, "syntax: check <version>\n")
		os.Exit(2)
	}

	v := os.Args[1]

	for _, n := range releasenotes.ReleaseNotes {
		if n.Version == v {
			return
		}
	}

	_, _ = fmt.Fprintf(os.Stderr, "update release note not found for version %v\n", v)
	os.Exit(1)
}
