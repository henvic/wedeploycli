package canceled

// Command skipped / canceled by the user
type Command struct {
	msg string
}

func (cc Command) Error() string {
	return cc.msg
}

// CancelCommand creates a 'cancelled command' error
// so the system can end the program with exit code 0
// when a user cancels a command on the CLI prompt
func CancelCommand(s string) error {
	return Command{
		msg: s,
	}
}
