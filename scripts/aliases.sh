#!/bin/bash

# Run code without building by invoking "l" (commented on purpose)
# A better alternative is to create a symbolic link to this file,
# like "make development-environment.sh" does
# e.g., "l link" instead of "lcp link"
# alias l="$GOPATH/src/github.com/henvic/wedeploycli/scripts/build-run.sh $@"

# Run go tests and generate test coverage for the current directory
alias gotest='go test -race -coverprofile=coverage.out && go tool cover -html coverage.out -o coverage.html'

# Open go tests coverage report for the current directory
alias goreport="open coverage.html"

# Check linting issues and exit with error code if there is any linting error
# Excluding vendor/ directory verification
alias golintt='test -z "$(golint `go list ./...` | tee /dev/stderr)"'

# Run govet
# Excluding vendor/ directory verification
alias govet='go vet -shadow $(go list ./...)'

# Start the Go docs server on port 6060 with playground enabled.
alias=godocserver='godoc -http :6060 -play'
