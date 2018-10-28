set _junit_report_content "<?xml version='1.0' encoding='utf-8'?>"
append _junit_report_content "<testsuites name='cli-functional-tests' tests='$_tests_total'"
append _junit_report_content " failures='$_tests_failed' errors='$_tests_errors'>"
append _junit_report_content $_junit_features_content
append _junit_report_content "</testsuites>"

add_to_junit_report $_junit_report_content

# report test counts
add_to_report "\nTotal number of tests: $_tests_total"
add_to_report "Total failed: $_tests_failed"
add_to_report "Total errors: $_tests_errors"

print_msg "Total number of tests: $_tests_total" green

if { $_tests_failed > 0 } {
  print_msg_stderr "Total failed: $_tests_failed failed tests :(" magenta
}

if { $_tests_errors > 0 } {
  print_msg_stderr "Total errors: $_tests_errors"
  exit 1
}

if { $_tests_failed == 0 } {
  print_msg "All tests passed! :D" green
}
