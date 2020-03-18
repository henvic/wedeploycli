package transport

import "github.com/henvic/wedeploycli/config"

// Settings for the transport.
type Settings struct {
	ConfigContext config.Context
	ProjectID     string
	Path          string
	WorkDir       string
}
