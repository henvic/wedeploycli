<h1 align="center">✨ CLI functional tests</h1>

 <h5 align="center">Functional tests for WeDeploy CLI using Expect/Tcl</h5>

## Dependencies
* Docker - to run infrastructure
* [Expect/Tcl](http://expect.sourceforge.net/)
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

## Running tests
* [If running the infra locally](https://github.com/wedeploy/training#development), add these lines to /etc/hosts.
```
127.0.0.1 api.wedeploy.xyz
127.0.0.1 git.wedeploy.xyz
```
* To run all tests
```
cd tests
TESTER_EMAIL={useremail} \
TESTER_PW={userpw} \
TEAMUSER_EMAIL={teamemail} \
TEAMUSER_PW={teampw} \
REMOTE=wedeploy.xyz \
./main.exp
```
* To run one test script, i.e. list.exp
```
cd tests
TESTER_EMAIL={useremail} \
TESTER_PW={userpw} \
TEAMUSER_EMAIL={teamemail} \
TEAMUSER_PW={teampw} \
REMOTE=wedeploy.xyz \
./main.exp -tclargs list.exp
```
* TESTER_EMAIL should have at least a standard plan and user will be created if it doesn't already exist.
* For any environment variable not provided, following default values are used:

| Variable       | Default value             |
| -------------  | ------------------------- |
| TEAMUSER_EMAIL | qa.team.user@wedeploy.com |
| TEAMUSER_PW    | L6P&ZExVXydC              |
| TESTER_EMAIL   | cli-tester@test.com       |
| TESTER_PW      | test                      |
| REMOTE         | wedeploy.xyz              |


## Test results
Results are reported in test-results/report.txt.  This contains a list of all scenarios from latest test run.  Any fails and errors encountered are listed below the respective scenario name.

Results are also compiled in junit format in test-results/TEST-report.xml.  

To view test reports on Jenkins CI
1. Click [here](https://ci.wedeploy.com/blue/organizations/jenkins/WeDeploy%2Fcli-functional-tests/activity/).  
1. Click on a test run.
1. For junit report, click 'Tests' in the top nav bar.
1. For report.txt, click 'Artifacts' in the top nav bar, then click functional/test-results/report.txt.

## Contributing
The tests are organized into Features and Scenarios.  A Feature is a test suite, which corresponds to a file in the functional/tests folder.  A Feature can contain one or more Scenarios.  Each Scenario is a test case.

Each Feature must begin with
```
Feature: {feature name}
```
and end with
```
TearDownFeature: {feature name}
```
Similarly, each Scenario must begin with
```
Scenario: {scenario name}
```
and end with
```
TearDownScenario: {scenario name}
```

Each scenario begins with setup steps.  Then the actual test steps are all enclosed within a `while` block.  The test steps are a series of `send` and `expect` commands.  In other words, `send` something to the command line, then `expect` something in the terminal output.  After the test steps are any necessary teardown steps.  Please see any of the test files for examples.

Here's some helpful quick start guides for using Expect/TCL:

- [Basic principles of using tcl-expect scripts](https://gist.github.com/Fluidbyte/6294378)
- [Expect and TCL mini reference manual](http://inguza.com/document/expect-and-tcl-mini-reference-manual)
