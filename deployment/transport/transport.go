package transport

import "github.com/wedeploy/cli/config"

// Settings for the transport.
type Settings struct {
	ConfigContext config.Context
	ProjectID     string
	Path          string
	WorkDir       string
}
