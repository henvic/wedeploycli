# WeDeploy CLI tool [![Build Status](https://travis-ci.com/wedeploy/cli.svg?token=a51FNuiJPYZtHhup9q1V&branch=master)](https://travis-ci.com/wedeploy/cli) [![Windows Build status](https://ci.appveyor.com/api/projects/status/06l69s8kc6nrqi74?svg=true)](https://ci.appveyor.com/project/wedeploy/cli) [![Coverage Status](https://coveralls.io/repos/wedeploy/cli/badge.svg)](https://coveralls.io/r/wedeploy/cli) [![codebeat badge](https://codebeat.co/badges/bd6acb49-ccdf-4045-a877-05da0198261a)](https://codebeat.co/projects/github-com-wedeploy-cli) [![Go Report Card](https://goreportcard.com/badge/github.com/wedeploy/cli)](https://goreportcard.com/report/github.com/wedeploy/cli) [![GoDoc](https://godoc.org/github.com/wedeploy/cli?status.svg)](https://godoc.org/github.com/wedeploy/cli)

Install this tool with

`curl https://cdn.wedeploy.com/cli/latest/wedeploy.sh -fsSL | bash`

or download from our [stable release channel](https://dl.equinox.io/wedeploy/we/stable).

To update this tool, just run `we update`.

## Dependencies
The following external soft dependencies are necessary to correctly run some commands:
* [docker](https://www.docker.com/)
* [git](https://git-scm.com/)

The availability of dependencies are tested just before its immediate use. If a required dependency is not found, an useful error message is printed and the calling process is terminated with an error code.

## Contributing
You can get the latest CLI source code with `go get -u github.com/wedeploy/cli`

The following commands are available and requires no arguments:

* **make development-environment**: install development environment for this project
* **make get-dependencies**: get versioned Go dependencies
* **make legal**: generate legal notices for existing dependencies
* **make list-packages**: list all Go packages of the project
* **make build**: compiles the program
* **make fast-test**: run all unit tests
* **make test**: run all unit and integration tests
* **make build-integration-tests**: generate integration tests suites
* **make release**: tag, build, and publish new version of the app
* **make promote**: publish version already released to a given distribution channel
* **make release-notes-page**: update release notes page on wedeploy.com

In lieu of a formal style guide, take care to maintain the existing coding style. Add unit tests for any new or changed functionality. Integration tests should be written as well.

## Committing and pushing changes
The master branch of this repository on GitHub is protected:
* force-push is disabled
* tests MUST pass on Travis before merging changes to master
* branches MUST be up to date with master before merging

Keep your commits neat and [well documented](https://wiki.openstack.org/wiki/GitCommitMessages). Try to always rebase your changes before publishing them.

## Maintaining code quality
[goreportcard](https://goreportcard.com/report/github.com/wedeploy/cli) can be used online or locally to detect defects and static analysis results from tools with a great overview.

Using go test and go cover are essential to make sure your code is covered with unit tests.

Check [scripts/aliases.sh](https://github.com/wedeploy/cli/tree/master/scripts/aliases.sh) for a list of aliases you might find useful for development / testing.

Always run `make test` before submitting changes.

## Integration testing
`make test` already runs the integration tests.

You can generate a test suite to run integration tests outside of your development environment (say, on a virtual machine).

Set the `$WEDEPLOY_CLI_INTEGRATION_TESTS_PATH` environment variable to save the integration test suites on a shared network volume you can access from each test machine to speed up development and debugging.

Build and distribute test suites for the supported systems:

```
$ make build-integration-tests
Building integration test suites for multiple platforms:
darwin...	wedeploy-cli-integration-darwin.test
linux...	wedeploy-cli-integration-linux.test
windows...	wedeploy-cli-integration-windows.test.exe
Warning: mocks directory placed on trash and recreated.

Integration test suites and its related mocks are saved in:
/vm-shared-network-volume/wedeploy-cli-integration-tests
```

And run it as a regular executable on each target system.

## Functional testing
See the [cli-functional-tests](http://github.com/wedeploy/cli-functional-tests) repository with tests written using [cucumber aruba](https://github.com/cucumber/aruba). Be aware functional tests might have drastic side-effects given they destroy data, so be sure you know what you are doing before you run them.

## Environment variables
See [envs/envs.go](envs/envs.go) for an up-to-date list of used environment variables.

## Hidden global flags
* `--defer-verbose` (`-V`) to defer verbose output until program termination
* `--no-verbose-requests` to avoid printing requests when using `--verbose`
* `--no-color` to print text without colors
