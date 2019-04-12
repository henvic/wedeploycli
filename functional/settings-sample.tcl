# Configuration for running the tests.

# debug mode: uncomment the line below
# exp_internal 1

# testing mode
# mode "basic" requires only a regular user account
# mode "complete" requires a user account + team user account (to change user plan)
set _mode "basic"

# create user if it doesn't exists
set _create_user false

# remote to use during tests
set _remote "wedeploy.com"
set _service_domain "wedeploy.sh"

# project prefix
set project_prefix ""

# account to run tests with
set _tester(email) "tester@example.com"
set _tester(password) "password"

# team user account to use
# this is only used if _mode is not default (0) or if _create_user is set to 1.
set _teamuser(email) "tester@example.com"
set _teamuser(password) "password"

# name of the binary file to test
set bin "liferay"
