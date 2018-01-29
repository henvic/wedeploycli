#!/bin/bash

# Run code without building by invoking "i"
# e.g., "i link" instead of "we link"
alias i="$GOPATH/src/github.com/wedeploy/cli/scripts/build-run.sh $@"

# Run go tests and generate test coverage for the current directory
alias gotest='go test -race -coverprofile=coverage.out && go tool cover -html coverage.out -o coverage.html'

# Open go tests coverage report for the current directory
alias goreport="open coverage.html"

# Check linting issues and exit with error code if there is any linting error
# Excluding vendor/ directory verification
alias golintt='test -z "$(golint ./... | grep -v "^vendor" | tee /dev/stderr)"'

# Run govet
# Excluding vendor/ directory verification
alias govet='go vet $(go list ./...)'
