# WeDeploy CLI tool [![Build Status](http://img.shields.io/travis/wedeploy/cli/master.svg?style=flat)](https://travis-ci.org/wedeploy/cli) [![Coverage Status](https://coveralls.io/repos/wedeploy/cli/badge.svg)](https://coveralls.io/r/wedeploy/cli) [![codebeat badge](https://codebeat.co/badges/bd6acb49-ccdf-4045-a877-05da0198261a)](https://codebeat.co/projects/github-com-wedeploy-cli) [![Go Report Card](https://goreportcard.com/badge/github.com/wedeploy/cli)](https://goreportcard.com/report/github.com/wedeploy/cli) [![GoDoc](https://godoc.org/github.com/wedeploy/cli?status.svg)](https://godoc.org/github.com/wedeploy/cli)

Install this tool with

`curl https://raw.githubusercontent.com/wedeploy/cli/master/install.sh -s | bash`

or download from our [stable release channel](https://dl.equinox.io/wedeploy/cli/stable).

To update this tool, just run `we update`.

## Dependencies
The following external soft dependencies are necessary to correctly run some commands:
* [docker](https://www.docker.com/)
* [git](https://git-scm.com/)

The availability of dependencies are tested just before its immediate use. If a required dependency is not found, an useful error message is printed and the calling process is terminated with an error code.

## Contributing
You can get the latest CLI source code with `go get -u github.com/wedeploy/cli`

In lieu of a formal style guide, take care to maintain the existing coding style. Add unit tests for any new or changed functionality. Integration tests should be written as well.

The master branch of this repository on GitHub is protected:
* force-push is disabled
* tests MUST pass on Travis before merging changes to master
* branches MUST be up to date with master before merging

Keep your commits neat. Try to always rebase your changes before publishing them.

[goreportcard](https://goreportcard.com/report/github.com/wedeploy/cli) can be used online or locally to detect defects and static analysis results from tools such as go vet, go lint, gocyclo, and more. Run [errcheck](https://github.com/kisielk/errcheck) to fix ignored error returns.

Using go test and go cover are essential to make sure your code is covered with unit tests.

Some aliases you might find useful for development / testing:

* Generating test coverage for the current directory: `alias gotest='go test -coverprofile=coverage.out && go tool cover -html coverage.out -o coverage.html'`
* Running code without building: `alias i="go run $HOME/projects/gocode/src/github.com/wedeploy/cli/main.go $1"` (i: development code, we: production binary)
* Opening coverage report: `alias goreport="open coverage.html"`
