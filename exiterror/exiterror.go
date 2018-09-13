package exiterror

// Error with exit code.
type Error struct {
	Err  string
	code int
}

// Error message.
func (e Error) Error() string {
	return e.Err
}

// Code for the exit syscall.
func (e Error) Code() int {
	return e.code
}

// New returns an error that formats as the given text with an associated exit code.
func New(text string, code int) Error {
	return Error{
		Err:  text,
		code: code,
	}
}
