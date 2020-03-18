package update

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/henvic/wedeploycli/config"
)

// Checker for update checks
type Checker struct {
	Cue chan error
}

// Check if an update is available on a goroutine
func (c *Checker) Check(ctx context.Context, conf *config.Config) {
	c.Cue = make(chan error, 1)
	go func() {
		c.Cue <- NotifierCheck(ctx, conf)
	}()
}

// Feedback of an update check
func (c *Checker) Feedback(conf *config.Config) {
	if c.Cue == nil {
		return
	}

	var err = <-c.Cue
	ue, ueok := err.(*url.Error)

	switch {
	case err == context.Canceled, ueok && ue.Err == context.Canceled:
		return
	case err != nil:
		_, _ = fmt.Fprintln(os.Stderr, "Update notification error:", err.Error())
		return
	default:
		Notify(conf)
	}
}
