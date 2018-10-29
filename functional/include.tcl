# List of scenario files that should be included (in order of execution).

set test_files {}

# override this list if -tclargs is used
if { $::argc > 1 } {
  foreach arg $::argv {
    if { $arg == "-tclargs" } { continue }
    lappend test_files $arg
  }

  return
}

# list of test files to run
lappend test_files \
  setup.exp \
  logout.exp \
  loggedout.exp \
  login.exp \
  loggedin.exp \
  list.exp \
  deploy.exp \
  log.exp \
  restart.exp \
  new.exp \
  delete.exp \
  domain.exp \
  env-var.exp \
  quota.exp \
  scale.exp \
  shell.exp \
  teardown.exp
