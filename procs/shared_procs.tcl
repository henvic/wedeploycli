#! /usr/bin/expect

proc add_to_report {text} {
  set file [open $::_test_report a+]
  puts $file $text
  close $file
}

proc control_c {} {
  send \003
  expect {
    timeout { handle_timeout; error "^C failed" }
    "cli-functional-tests/"
  }
}

proc expectation_not_met {message} {
  print_msg "Expectation not met: $message" red
  set stack [print_stack]
  add_to_report "Expectation Not Met Error: $message\n$stack"
  set timeout $::_default_timeout
}

proc handle_timeout {{message ""}} {
  print_msg "Timeout Error: $message" red
  set stack [print_stack]

  add_to_report "  Timeout Error: $message\n$stack"

  set timeout $::_default_timeout
  control_c
}

proc print_msg {text {color cyan}} {
  if { [string match {SCENARIO:*} $text] } {
    set color magenta
    add_to_report "\n$text"
  }

  if { [string match {Finished!} $text] } { set color green }

  switch $color {
    green { set color_code 32 }
    magenta { set color_code 35 }
    red { set color_code 31 }
    cyan -
    default { set color_code 36}
  }

  puts "\n\033\[01;$color_code;m$text \033\[0;m\n"
}

proc print_stack {} {
  set stack_size [info frame]
  set stack_payload_size [expr {$stack_size - 3}]
  set stack {}

  for {set frame_index $stack_payload_size} {$frame_index >= 1} {incr frame_index -1} {
    set frame [info frame $frame_index]
    set cmd [dict get $frame cmd]
    set file -
    set line -
    if {[dict exists $frame file]} {set file [dict get $frame file]}
    if {[dict exists $frame line]} {set line [dict get $frame line]}

    set max_string_size 30
    if { [string length $cmd] > $max_string_size} {
      set cmd "[string range $cmd 0 $max_string_size]..."
    }

    set stack_line "  ERROR: [file tail $file] | line $line | cmd: $cmd "

    lappend stack $stack_line
  }

  puts [join $stack "\n"]
  return [join $stack "\n"]
}

proc login {email pw} {
  send "we login --no-browser\r"
  expect {
    timeout { handle_timeout; error "Login failed" }
    "Your email:" {
      send "$email\r"
      expect "Now, your password:"
      send "$pw\r"
      expect {
        timeout { handle_timeout; error "Login failed" }
        "Authentication failed" { error "Login failed" }
        "Type a command and press Enter to execute it."
      }
    }
    "Already logged in"
  }
}

proc logout {email} {
  send "we logout\r"
  expect {
    timeout { handle_timeout }
    "You are not logged in" {}
    "You ($email) have been logged out"
  }
}
