#!/usr/bin/expect

proc Feature: {name} {
  set ::_current_feature "$name"
  print_msg "FEATURE: $name" magenta
  add_to_report "\nFEATURE: $name"
  set ::_scenarios_count 0
  set ::_tests_errors_by_feature 0
  set ::_tests_failed_by_feature 0
  set ::_junit_scenarios_content ""
  set ::_time_by_feature [clock milliseconds]
}

proc end_feature {name} {
  set end [clock milliseconds]
  set time [expr {$end - $::_time_by_feature}]
  append ::_junit_features_content "<testsuite hostname='localhost' id='$name' name='$name' tests='$::_scenarios_count' time='$time' errors='$::_tests_errors_by_feature' failures='$::_tests_failed_by_feature'>"
  append ::_junit_features_content $::_junit_scenarios_content
  append ::_junit_features_content "</testsuite>"
  print_msg "COMPLETED FEATURE: $name in $time milliseconds" magenta
  add_to_report "\nCOMPLETED FEATURE: $name in $time milliseconds"
}

proc begin_scenario {name} {
  set ::_current_scenario "$name"
  set ::_junit_scenarios_error_content ""
  incr ::_tests_total 1
  print_msg "SCENARIO: $name" magenta
  add_to_report "\nSCENARIO: $name"
  incr ::_scenarios_count 1
  set ::_time_by_scenario [clock milliseconds]
}

proc end_scenario {name} {
  set end [clock milliseconds]
  set time [expr {$end - $::_time_by_scenario}]
  append ::_junit_scenarios_content "<testcase id='$name' name='$name' time='$time'>"
  append ::_junit_scenarios_content $::_junit_scenarios_error_content
  append ::_junit_scenarios_content "</testcase>"

  print_msg "\nCOMPLETED SCENARIO: $name in $time milliseconds" magenta
  add_to_report "\nCOMPLETED SCENARIO: $name in $time milliseconds"
}

proc Scenario: {name script {teardown ""}} {
  begin_scenario $name

  while {1} {
    eval $script
    break
  }

  if { $teardown != "" } {
    print_msg "Teardown in progress..."
    eval $teardown
    print_msg "Teardown complete"
  }

  end_scenario $name
}

# place - in front of Scenario to skip it
proc -Scenario: {name script {teardown ""}} {
}

proc create_report {} {
  # cleanup/initialize test report file
  set report_file [open $::_test_report w]
  close $report_file
}

proc add_to_report {text} {
  set file [open $::_test_report a+]
  puts $file $text
  close $file
}

proc add_to_junit_report {text} {
  set file [open $::_junit_test_report w]
  puts $file $text
  close $file
}

proc control_c {} {
  send \003
  expect {
    timeout { error "^C failed" }
    "$::_root_dir"
  }
}

proc exit_shell {} {
  send "exit\r"
  expect {
    timeout { error "Failed to exit shell" }
    "$::_root_dir"
  }
}

proc expectation_not_met {message} {
  incr ::_tests_failed 1
  incr ::_tests_failed_by_feature 1
  print_msg_stderr "Expectation not met: $message"
  set stack [print_stack]
  add_to_report "Expectation Not Met Error: $message\n$stack"
  set timeout $::_default_timeout
  append ::_junit_scenarios_error_content "<failure>Expectation Not Met Error: $message\n$stack</failure>"
}

proc handle_timeout {{message ""}} {
  incr ::_tests_failed 1
  incr ::_tests_failed_by_feature 1
  print_msg_stderr "Timeout Error: $message"
  set stack [print_stack]
  add_to_report "Timeout Error: $message\n$stack"
  append ::_junit_scenarios_error_content "<failure>Timeout Error: $message\n$stack</failure>"
  set timeout $::_default_timeout
  control_c
}

proc print_msg {text {color cyan}} {
  puts [get_msg $text $color]
}

proc print_msg_stderr {text {color red}} {
  puts stderr [get_msg $text $color]
}

proc get_msg {text color} {
  switch $color {
    green { set color_code 32 }
    magenta { set color_code 35 }
    red { set color_code 31 }
    cyan -
    default { set color_code 36 }
  }

  return "\033\[01;$color_code;m$text \033\[0;m\n"
}

proc print_stack {} {
  set stack_size [info frame]
  set stack_payload_size [expr {$stack_size - 3}]
  set stack {}

  for { set frame_index $stack_payload_size } { $frame_index >= 1 } { incr frame_index -1 } {
    set frame [info frame $frame_index]
    set cmd [dict get $frame cmd]
    set file -
    set line -
    if { [dict exists $frame file] } { set file [dict get $frame file] }
    if { [dict exists $frame line] } { set line [dict get $frame line] }

    set max_string_size 120
    if { [string length $cmd] > $max_string_size } {
      set cmd "[string range $cmd 0 $max_string_size]..."
    }

    set stack_line "[file tail $file], line $line\n  $cmd"

    lappend stack $stack_line
  }

  puts [join $stack "\n"]
  return [join $stack "\n"]
}

proc prefixed_project {project} {
  return ${::project_prefix}$project
}

proc delete_projects {projects} {
  foreach project $projects {
    exec $::bin delete -p [prefixed_project $project] --force
  }
}

proc login {email password} {
  spawn $::bin login --no-browser
  expect {
    timeout {
      print_msg_stderr "\nLogin timed out\n"
      exit 1
    }
    "Your email:" {
      send "$email\r"
      expect "Now, your password:"
      send "$password\r"
      expect {
        timeout {
          print_msg_stderr "\nLogin timed out\n"
          exit 1
        }
        "Authentication failed" {
          print_msg_stderr "\nLogin failed\n"
          exit 1
        }
        "Type a command and press Enter to execute it."
      }
    }
    "Already logged in"
  }
}

proc logout {email} {
  send "$::bin logout\r"
  expect {
    timeout {
      print_msg_stderr "\nLogout timed out\n"
    }
    "You are not logged in" {}
    "You ($email) have been logged out"
  }
}
