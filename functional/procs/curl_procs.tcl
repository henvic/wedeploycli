#! /usr/bin/expect

package require TclCurl

set base_url "https://api.$::_remote"
set auth $::_tester(email):$::_tester(pw)
set team_auth $::_teamuser(email):$::_teamuser(pw)
set auth_header {"Authorization: Bearer token"}
set content_type_header {"Content-Type: application/json; charset=utf-8"}

proc http_get {url userpw {args}} {
  if { [llength $args] > 0 } {
    set pairs {}
    foreach {name value} $args {
      lappend pairs "[curl::escape $name]=[curl::escape $value]"
    }
    append url ? [join $pairs &]
  }

  set curl_handle [curl::init]
  $curl_handle configure \
    -url $url \
    -userpwd $userpw \
    -bodyvar body

  if { [catch {$curl_handle perform} curl_error_number] } {
    error [curl::easystrerror $curl_error_number]
  }

  set code [$curl_handle getinfo httpcode]
  $curl_handle cleanup

  return [list $code $body]
}

proc http_post {url userpw data} {
  set curl_handle [curl::init]
  $curl_handle configure \
    -url $url \
    -userpwd $userpw \
    -httpheader $::content_type_header \
    -post 1 \
    -postfields $data \
    -bodyvar body

  if { [catch {$curl_handle perform} curl_error_number] } {
    error [curl::easystrerror $curl_error_number]
  }

  set code [$curl_handle getinfo httpcode]
  $curl_handle cleanup

  return [list $code $body]
}

proc handle_response {message body} {
  append message "\n  $body"
  add_to_report "$message"
  print_msg $message red
}

proc assert_service_exists {project service} {
  print_msg "Verifying service $service-$project"

  set url $::base_url/projects/$project/services/$service
  set response [http_get $url $::auth]
  set response_code [lindex $response 0]
  set body [lindex $response 1]

  if { $response_code != 200 } {
    expectation_not_met "Could not verify service $service-$project\n$body"
  }
}

proc create_project {project {env false}} {
  print_msg "Creating project $project"

  set timeout 30
  set url $::base_url/projects
  set data "\{\"projectId\":\"$project\", \"environment\": $env\}"
  set response [http_post $url $::auth $data]
  set response_code [lindex $response 0]
  set body [lindex $response 1]
  set timeout $::_default_timeout

  if { $response_code != 200 } {
    handle_response "Project $project could not be created" $body
  }
}

proc create_service {project service {image wedeploy/hosting}} {
  print_msg "Creating service $service for project $project"

  set timeout 30
  set url $::base_url/projects/$project/services
  set data "\{\"serviceId\":\"$service\",\"image\":\"$image\"\}"
  set response [http_post $url $::auth $data]
  set response_code [lindex $response 0]
  set body [lindex $response 1]
  set timeout $::_default_timeout

  if { $response_code != 200 } {
    handle_response "Service $service could not be created" $body
  }
}

proc create_user {email {pw test} {name Tester} {plan standard}} {
  print_msg "Creating user $email"

  set data "\{\
      \"email\": \"$email\",\
      \"password\": \"$pw\",\
      \"name\": \"$name\"\}"
  set url $::base_url/user/create
  set response [http_post $url $::team_auth $data]
  set response_code [lindex $response 0]
  set body [lindex $response 1]

  if { $response_code != 200 } {
    handle_response "Could not create user $email" $body
    error "Error creating user"
  }

  # get token and confirm user
  regexp {"confirmed":"(.*?)"} $body matched confirm_token

  set params "email $email confirmationToken $confirm_token"
  set response [http_get $::base_url/confirm $::auth {*}$params]
  set response_code [lindex $response 0]
  set body [lindex $response 1]

  if { $response_code != 302 } {
    handle_response "Could not confirm user $email" $body
    error "Error confirming email"
  }

  set_user_plan $plan
}

proc delete_project {project} {
  print_msg "Deleting project $project"

  set url $::base_url/projects/$project

  set curl_handle [curl::init]
  $curl_handle configure \
    -customrequest DELETE \
    -url $url \
    -userpwd $::auth

  if { [catch {$curl_handle perform} curl_error_number] } {
    error [curl::easystrerror $curl_error_number]
  }

  set code [$curl_handle getinfo httpcode]
  $curl_handle cleanup

  if { $code != 204 } {
    handle_response "Could not delete project $project" ""
  }
}

proc get_container_ids {project service} {
  set url $::base_url/projects/$project/services/$service/instances
  set response [http_get $url $::auth]
  set response_code [lindex $response 0]
  set body [lindex $response 1]

  if { $response_code != 200 } {
    handle_response "Could not get container ids" $body
    error "Could not get container ids"
  }

  set match_list [regexp -all -inline {"containerId":"(.*?)"} $body]
  set ids {}

  foreach { whole container_id } $match_list {
    lappend ids $container_id
  }

  return $ids
}

# get user id  of currently logged in user, presumed to be $::_tester(email)
proc get_user_id {} {
  set url $::base_url/user
  set response [http_get $url $::auth]
  set response_code [lindex $response 0]
  set body [lindex $response 1]

  if { $response_code != 200 } {
    handle_response "Could not get user id" $body
  }

  regexp {"id":"(.*?)"} $body matched user_id
  return $user_id
}

# update plan of logged in user
proc set_user_plan {plan} {
  print_msg "Setting user plan to $plan"

  set data "\{\"planId\": \"$plan\"\}"
  set user_id [get_user_id]
  set url $::base_url/admin/users/$user_id

  set curl_handle [curl::init]
  $curl_handle configure \
    -customrequest PATCH \
    -url $url \
    -userpwd $::team_auth \
    -httpheader $::content_type_header \
    -postfields $data \
    -bodyvar body

  if { [catch {$curl_handle perform} curl_error_number] } {
    error [curl::easystrerror $curl_error_number]
  }

  set response_code [$curl_handle getinfo httpcode]
  $curl_handle cleanup

  if { $response_code != 200 } {
    handle_response "Could not update user plan" $body
  }
}

proc user_exists {email} {
  set url $::base_url/admin/users
  set response [http_get $url $::team_auth]
  set response_code [lindex $response 0]
  set body [lindex $response 1]

  if { $response_code != 200 } {
    handle_response "Could not get users" $body
    error "Could not get users"
  }

  return [string match *$email* $body]
}
