# set remote
exec we curl enable
exec we remote set qa-remote $_remote
exec we remote default qa-remote

create_report

# create tester if user doesn't already exist
if {$_create_user} {
  create_user
}

# print we version
puts [exec we version]

# print list of remotes
set result [exec we remote]
puts "we remote\n$result"

proc create_user {} {
  if { [user_exists $::_tester(email)] } {
    return
  }

  if { [catch { create_user $::_tester(email) }] } {
    puts stderr "Error creating test user! Aborting tests."
    exit 1
  }
}