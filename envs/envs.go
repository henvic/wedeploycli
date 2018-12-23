package envs

const (
	// TZ is used to change the timezone for the program
	TZ = "TZ"

	// CustomHome to use instead of $HOME
	CustomHome = "WEDEPLOY_CUSTOM_HOME"

	// UnsafeVerbose enables verbose mode, including showing tokens and so on
	UnsafeVerbose = "WEDEPLOY_UNSAFE_VERBOSE"

	// GitCredentialRemoteToken is the environment variable used for git credential-helper
	GitCredentialRemoteToken = "WEDEPLOY_REMOTE_TOKEN"

	// MachineFriendly is used for returning more consistent output
	// TODO(henvic): consider adopting the NO_COLOR standard (see no-color.org)
	MachineFriendly = "WEDEPLOY_MACHINE_FRIENDLY"

	// SkipTerminalVerification makes isTerm.Check() return true always
	SkipTerminalVerification = "WEDEPLOY_SKIP_TERMINAL_VERIFICATION"

	// SkipTLSVerification is used to skip the TLS/SSL verification
	SkipTLSVerification = "WEDEPLOY_SKIP_TLS_VERIFICATION"
)
