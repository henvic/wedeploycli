package update

import "github.com/wedeploy/cli/config"

// Checker for update checks
type Checker struct {
	Cue chan error
}

// Check if an update is available on a goroutine
func (c *Checker) Check(conf *config.Config) {
	c.Cue = make(chan error, 1)
	go func() {
		c.Cue <- NotifierCheck(conf)
	}()
}

// Feedback of an update check
func (c *Checker) Feedback(conf *config.Config) {
	if c.Cue == nil {
		return
	}

	var err = <-c.Cue
	switch err {
	case nil:
		Notify(conf)
	default:
		println("Update notification error:", err.Error())
	}
}
