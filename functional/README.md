# ✨ CLI functional tests

This is the test suite for the functional tests for the Liferay Cloud Platform CLI. The functional testing suite is written with [TCL](https://tcl.tk/) and uses the [expect](https://core.tcl.tk/expect) toolkit.

## Dependencies

* [Expect/Tcl](https://expect.sourceforge.net/)
* TclCurl package
    1. Install [ActiveTcl](https://www.activestate.com/activetcl/downloads)
    1. Clone [tclcurl-fa](https://github.com/flightaware/tclcurl-fa) and follow instructions to make and install
    1. Copy `tclcurl.tcl` from tclcurl-fa/generic to tclcurl-fa
    1. Place entire tclcurl-fa folder in `$auto_path`.  Check `$auto_path` by typing in the terminal:
        ```
        tclsh
        % puts $auto_path
        ```
    1. Confirm installation
        ```
        % package require TclCurl
        7.22.0
        ```

In addition, if you are running the infrastructure locally, [Docker](https://www.docker.com/) and some [assembly](https://github.com/wedeploy/training#development) is required.

## Running tests
You can run the tests with `make functional-tests`.

Copy the settings-sample.tcl to settings.tcl and configure it accordingly.

## Test results
Reports of all scenarios that were run at the latest test are saved in the tests/results/report.txt file.

Results are also compiled in [JUnit XML format](https://wiki.jenkins.io/display/JENKINS/JUnit+Plugin) in tests/results/TEST-report.xml.  

To view test reports on Jenkins CI open the [cli-functional-tests Jenkins pipeline](https://ci.wedeploy.com/blue/organizations/jenkins/WeDeploy%2Fcli-functional-tests/activity/).

From there you can select a test run. For JUnit reports, click "Tests" in the top nav bar. For reports.txt, click "Artifacts" in the top nav bar, then select "functional/results/report.txt".

## Contributing
The tests are organized into Features and Scenarios.  A Feature is a test suite, which corresponds to a file in the functional/tests folder.  A Feature may contain one or more Scenarios. Each Scenario is a test case.

Each test file should begin with
```
Feature: {feature name}
```
One or more scenarios follow with this structure:
```
Scenario: {scenario name} {
    setup steps
    test step 1
    test step 2
    ...
} { teardown steps }
```

Each scenario begins with setup steps. The test steps are a series of `send` and `expect` commands. In other words, `send` something to the command line, then `expect` something in the terminal output. The last `{}` block contains the teardown steps. It may be omitted if there is no teardown. Please see any of the test files for examples.

You can also use `autoexpect` to help you creating tests.

# References
* [Basic principles of using tcl-expect scripts](https://gist.github.com/Fluidbyte/6294378)
* [Expect and TCL mini reference manual](http://inguza.com/document/expect-and-tcl-mini-reference-manual)
* [Tcler's Wiki](https://wiki.tcl-lang.org/)
