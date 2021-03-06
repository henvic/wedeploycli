# About this repository
The code you find here is legacy and was originally created for WeDeploy (that originated Liferay Cloud Platform later on). This code is not maintained or used anymore, but there are parts of it that might be useful for many Go projects (especially CLI tools).

To quickly test it, you can run:

```shell
go install github.com/henvic/wedeploycli/cmd/lcp
```

This code was previously at github.com/wedeploy/cli.

# Liferay Cloud Platform CLI tool [![Build Status](https://travis-ci.org/henvic/wedeploycli.svg?branch=master)](https://travis-ci.org/henvic/wedeploycli) [![Coverage Status](https://coveralls.io/repos/henvic/wedeploycli/badge.svg)](https://coveralls.io/r/henvic/wedeploycli) [![codebeat badge](https://codebeat.co/badges/bd6acb49-ccdf-4045-a877-05da0198261a)](https://codebeat.co/projects/github-com-wedeploy-cli) [![Go Report Card](https://goreportcard.com/badge/github.com/henvic/wedeploycli)](https://goreportcard.com/report/github.com/henvic/wedeploycli) [![GoDoc](https://godoc.org/github.com/henvic/wedeploycli?status.svg)](https://godoc.org/github.com/henvic/wedeploycli)

**The contents below are outdated.**

Install the tool with:

```shell
curl https://cdn.liferay.cloud/cli/latest/lcp.sh -fsSL | bash
```

or download from our [stable release channel](https://dl.equinox.io/wedeploy/lcp/stable).

To update this tool, just run `lcp update`.

## Dependencies
The following external soft dependencies are necessary to correctly run some commands:

* [docker](https://www.docker.com/)
* [git](https://git-scm.com/)

The availability of dependencies are tested just before its immediate use. If a required dependency is not found, an error message is printed and the calling process is terminated with an error code.

## Contributing
You can get the latest CLI source code with `go get -u github.com/henvic/wedeploycli`. Alternatively, clone the repo as usual. 

The following commands are available and requires no arguments:

* **make clean**: clears generated binaries
* **make development-environment**: install development environment for this project
* **make legal**: generate legal notices for existing dependencies
* **make list-packages**: list all Go packages of the project
* **make build**: compiles the program
* **make fast-test**: run all unit tests
* **make test**: run all unit and integration tests
* **make functional-tests**: run functional tests
* **make build-integration-tests**: generate integration tests suites
* **make release**: tag, build, and publish new version of the app
* **make promote**: publish version already released to a given distribution channel
* **make release-notes-page**: update release notes page

In lieu of a formal style guide, take care to maintain the existing coding style. Add unit tests for any new or changed functionality. Integration tests should be written as well.

## Committing and pushing changes
The master branch of this repository on GitHub is protected:
* force-push is disabled
* tests MUST pass on Travis before merging changes to master
* branches MUST be up to date with master before merging

Keep your commits neat and [well documented](https://wiki.openstack.org/wiki/GitCommitMessages). Try to always rebase your changes before publishing them.

## Maintaining code quality
[goreportcard](https://goreportcard.com/report/github.com/henvic/wedeploycli) can be used online or locally to detect defects and static analysis results from tools with a great overview.

Using go test and go cover are essential to make sure your code is covered with unit tests.

Check [scripts/aliases.sh](https://github.com/henvic/wedeploycli/tree/master/scripts/aliases.sh) for a list of aliases you might find useful for development / testing.

Always run `make test` before submitting changes.

## Integration testing
`make test` already runs the integration tests.

You can generate a test suite to run integration tests outside of your development environment (say, on a virtual machine).

Set the `$WEDEPLOY_CLI_INTEGRATION_TESTS_PATH` environment variable to save the integration test suites on a shared network volume you can access from each test machine to speed up development and debugging.

Build and distribute test suites for the supported systems:

```
$ make build-integration-tests
Building integration test suites for multiple platforms:
darwin...	lcp-cli-integration-darwin.test
linux...	lcp-cli-integration-linux.test
windows...	lcp-cli-integration-windows.test.exe
Warning: mocks directory placed on trash and recreated.

Integration test suites and its related mocks are saved in:
/vm-shared-network-volume/lcp-cli-integration-tests
```

And run it as a regular executable on each target system.

## Functional testing
Functional tests for the CLI are written in the [Tcl](https://tcl.tk/) programming language and uses [Expect](https://core.tcl.tk/expect/). See the [README](functional/README.md) at the functional directory.

These tests are run by connecting to a real Liferay Cloud infrastructure, therefore caution to avoid data loss is advised. For this very reason it refuses to be run on a non-empty user account by default.

You need to copy `functional/settings-sample.tcl` to `functional/settings.tcl` and configure it appropriately.

## Environment variables
See [envs/envs.go](envs/envs.go) for an up-to-date list of used environment variables.

## Hidden global flags
* `--defer-verbose` (`-V`) to defer verbose output until program termination
* `--no-verbose-requests` to avoid printing requests when using `--verbose`
* `--no-color` to print text without colors
