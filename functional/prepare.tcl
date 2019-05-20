# set remote
print_msg_stderr "enabling \"lcp curl\""
exec $::bin curl enable

print_msg_stderr "creating qa-remote $_remote"
exec $::bin remote set qa-remote $_remote

print_msg_stderr "set default remote to qa-remote"
exec $::bin remote default qa-remote

create_report

# create tester if user doesn't already exist
if {$_create_user} {
  if { [user_exists $::_tester(email)] } {
    return
  }

  if { [catch { create_user $::_tester(email) }] } {
    puts stderr "Error creating test user! Aborting tests."
    exit 1
  }
}

# print $::bin version
puts [exec $::bin version]

# print list of remotes
set result [exec $::bin remote]
puts "$::bin remote\n$result"
