# Go API Client for WeDeploy Project [![Build Status](https://travis-ci.com/wedeploy/wedeploy-sdk-go.svg?token=a51FNuiJPYZtHhup9q1V&branch=master)](https://travis-ci.com/wedeploy/wedeploy-sdk-go) [![Coverage Status](https://coveralls.io/repos/wedeploy/wedeploy-sdk-go/badge.svg)](https://coveralls.io/r/wedeploy/wedeploy-sdk-go) [![codebeat badge](https://codebeat.co/badges/20267bde-2111-4bdf-a101-78d062a7cc99)](https://codebeat.co/projects/github-com-wedeploy-wedeploy-sdk-go) [![Go Report Card](https://goreportcard.com/badge/github.com/wedeploy/wedeploy-sdk-go)](https://goreportcard.com/report/github.com/wedeploy/wedeploy-sdk-go) [![GoDoc](https://godoc.org/github.com/wedeploy/wedeploy-sdk-go?status.svg)](https://godoc.org/github.com/wedeploy/wedeploy-sdk-go)

To install:

`
go get -u github.com/wedeploy/wedeploy-sdk-go
`

## Contributing
You can get the latest wedeploy-sdk-go source code with `go get -u github.com/wedeploy/wedeploy-sdk-go`

In lieu of a formal style guide, take care to maintain the existing coding style. Add unit tests for any new or changed functionality. Integration tests should be written as well.

The master branch of this repository on GitHub is protected:
* force-push is disabled
* tests MUST pass on Travis before merging changes to master
* branches MUST be up to date with master before merging

Keep your commits neat. Try to always rebase your changes before publishing them.

[goreportcard](https://goreportcard.com/report/github.com/wedeploy/wedeploy-sdk-go) can be used online or locally to detect defects and static analysis results from tools such as go vet, go lint, gocyclo, and more. Run [errcheck](https://github.com/kisielk/errcheck) to fix ignored error returns.

Using go test and go cover are essential to make sure your code is covered with unit tests.

## License

[BSD-3-Clause](https://spdx.org/licenses/BSD-3-Clause.html), © Liferay, Inc.
