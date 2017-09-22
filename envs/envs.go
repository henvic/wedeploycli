package envs

const (
	// CustomHome to use instead of $HOME
	CustomHome = "WEDEPLOY_CUSTOM_HOME"

	// UnsafeVerbose enables verbose mode, including showing tokens and so on
	UnsafeVerbose = "WEDEPLOY_UNSAFE_VERBOSE"

	// GitCredentialRemoteToken is the environment variable used for git credential-helper
	GitCredentialRemoteToken = "WEDEPLOY_REMOTE_TOKEN"

	// MachineFriendly is used for returning more consistent output
	MachineFriendly = "WEDEPLOY_MACHINE_FRIENDLY"

	// SkipTerminalVerification makes isTerm.Check() return true always
	SkipTerminalVerification = "WEDEPLOY_SKIP_TERMINAL_VERIFICATION"
)
