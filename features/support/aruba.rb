require 'aruba/cucumber'

Before do
	aruba.config.activate_announcer_on_command_failure = [:stdout]
end