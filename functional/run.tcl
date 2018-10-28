# run tests
foreach test_file $test_files {
  if { [catch { source tests/$test_file } errmsg] } {
    incr ::_tests_errors 1
    incr ::_tests_errors_by_feature 1
    set message "\nUnexpected error occurred:\n"
    append message "ErrorMsg: $errmsg\n"
    append message "ErrorCode: $errorCode\n"
    append message "ErrorInfo:\n$errorInfo\n"
    puts stderr $message
    add_to_report $message
    append _junit_scenarios_error_content "<error>$message</error>"

    end_scenario $_current_scenario
  }

  end_feature $_current_feature
}
