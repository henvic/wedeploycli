#! /usr/bin/expect

source ../procs/shared_procs.tcl

package require TclCurl

set base_url https://api.wedeploy.xyz
set auth $::_tester(email):$::_tester(pw)
set auth_header {"Authorization: Bearer token"}
set content_type_header {"Content-Type: application/json; charset=utf-8"}

proc http_get {url {args}} {
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
		-userpwd $::auth

	catch { $curl_handle perform } curl_error_number
	if { $curl_error_number != 0 } {
		error [curl::easystrerror $curl_error_number]
	}

	set code [$curl_handle getinfo httpcode]
	$curl_handle cleanup

	return $code
}

proc http_post {url userpw data} {
	set curl_handle [curl::init]
	$curl_handle configure \
		-url $url \
		-userpwd $userpw \
		-httpheader $::content_type_header \
		-post 1 \
		-postfields $data

	catch { $curl_handle perform } curl_error_number
	if { $curl_error_number != 0 } {
		error [curl::easystrerror $curl_error_number]
	}

	set code [$curl_handle getinfo httpcode]
	$curl_handle cleanup

	return $code
}

proc create_project {project} {
	print_msg "Creating project $project"

	set timeout 30
	set url $::base_url/projects
	set data "\{\"projectId\":\"$project\"\}"
	set response_code [http_post $url $::auth $data]
	set timeout $::_default_timeout

	if { $response_code != 200 } {
		set message "Project $project could not be created"
		add_to_report $message
		print_msg $message red
	}
}

proc create_service {project service {image wedeploy/hosting}} {
	print_msg "Creating service $service for project $project"

	set timeout 30
	set url $::base_url/projects/$project/services
	set data "\{\"serviceId\":\"$service\",\"image\":\"$image\"\}"
	set response_code [http_post $url $::auth $data]
	set timeout $::_default_timeout

	if { $response_code != 200 } {
		set message "Service $service could not be created"
		add_to_report $message
		print_msg $message red
	}
}

proc create_user {email {pw test} {name Tester} {plan standard}} {
	print_msg "Creating user $email"

	set url localhost:8082/users
	set data "\{\
				\"confirmed\": null,\
				\"email\": \"$email\",\
				\"password\": \"$pw\",\
				\"name\": \"$name\",\
				\"planId\": \"$plan\"\}"

	set curl_handle [curl::init]
	$curl_handle configure \
		-url $url \
		-httpheader $::auth_header \
		-httpheader $::content_type_header \
		-post 1 \
		-postfields $data

	catch { $curl_handle perform } curl_error_number
	if { $curl_error_number != 0 } {
		error [curl::easystrerror $curl_error_number]
	}

	set code [$curl_handle getinfo httpcode]
	$curl_handle cleanup

	if { $code != 200 } {
		error "Could not create user $email"
	}
}

proc delete_project {project} {
	print_msg "Deleting project $project"

	set url $::base_url/projects/$project

	set curl_handle [curl::init]
	$curl_handle configure \
		-customrequest DELETE \
		-url $url \
		-userpwd $::auth

	catch { $curl_handle perform } curl_error_number
	if { $curl_error_number != 0 } {
		error [curl::easystrerror $curl_error_number]
	}

	set code [$curl_handle getinfo httpcode]
	$curl_handle cleanup

	if { $code != 204 } {
		set message "Could not delete project $project"
		add_to_report $message
		print_msg $message red
	}
}

proc verify_service_exists {project service} {
	set url $::base_url/projects/$project/services/$service
	set response_code [http_get $url]

	if { $response_code != 200 } {
		set message "Project $project with service $service doesn't exist"
		add_to_report $message
		print_msg $message red
	}
}