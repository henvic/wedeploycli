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
REMOTE=wedeploy.xyz \
./main.exp
```
* To run one test script, i.e. list.exp
```
cd tests
TESTER_EMAIL={useremail} \
TESTER_PW={userpw} \
REMOTE=wedeploy.xyz \
./main.exp -tclargs list.exp
```
* User with TESTER_EMAIL is assumed to exist, and should have at least a standard plan.  If the environment variables are not provided, following default values are used:

| Variable      | Default value       |
| ------------- | ------------------- |
| TESTER_EMAIL  | cli-tester@test.com |
| TESTER_PW     | test                |
| REMOTE        | wedeploy.xyz        |


## Test results
Results are reported in test-results/report.txt.  This contains a list of all scenarios from latest test run.  Any errors encountered are listed below the respective scenario name.
