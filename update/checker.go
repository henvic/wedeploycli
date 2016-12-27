package update

// Checker for update checks
type Checker struct {
	Cue chan error
}

// Check if an update is available on a goroutine
func (c *Checker) Check() {
	c.Cue = make(chan error, 1)
	go func() {
		c.Cue <- NotifierCheck()
	}()
}

// Feedback of an update check
func (c *Checker) Feedback() {
	if c.Cue == nil {
		return
	}

	var err = <-c.Cue
	switch err {
	case nil:
		Notify()
	default:
		println("Update notification error:", err.Error())
	}
}
