package canceled

// Command skipped / canceled by the user
type Command struct {
	msg string

	quiet bool
}

func (cc Command) Error() string {
	return cc.msg
}

// Quiet tells whether the error message should be print on termination
func (cc Command) Quiet() bool {
	return cc.quiet
}

// CancelCommand creates a 'cancelled command' error
// so the system can end the program with exit code 0
// when a user cancels a command on the CLI prompt
func CancelCommand(s string) error {
	return Command{
		msg: s,
	}
}

// Skip creates a 'quietly cancelled command'
func Skip() error {
	return Command{
		quiet: true,
	}
}
