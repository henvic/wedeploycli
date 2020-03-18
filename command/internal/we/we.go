package we

import "github.com/henvic/wedeploycli/config"

var ctx config.Context

// Context gets the context of the application global state
func Context() config.Context {
	return ctx
}

// WithContext sets the context for the application global state
func WithContext(c *config.Context) {
	ctx = *c
}
