#!/bin/bash
set -euox pipefail

# TODO(henvic): install specific versions of the commands
# when the 3 latest releases of the Go toolchains supports it using @tag.
go get github.com/mattn/goveralls
go get golang.org/x/lint/golint
go get honnef.co/go/tools/cmd/staticcheck
go get github.com/securego/gosec/cmd/gosec
go get golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
