set settingsError 0

foreach var {
  _mode \
  _create_user \
  _remote \
  _teamuser \
  _teamuser(email) \
  _teamuser(password) \
  _tester \
  _tester(email) \
  _tester(password)} {
  if {![info exists $var]} {
    puts stderr "Settings variable \"$var\" is not set."
    set settingsError 1
  }
}

if {$settingsError == 1} {
  puts stderr "config issue: fix the reported issues and try again."
  exit 1
}
