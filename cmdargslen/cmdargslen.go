package cmdargslen

import (
	"errors"

	"github.com/spf13/cobra"
)

var (
	errTooFew  = errors.New("Invalid number of arguments: too few")
	errTooMany = errors.New("Invalid number of arguments: too many")
)

// Validate the number of arguments
func Validate(args []string, min, max int) error {
	switch {
	case len(args) < min:
		return errTooFew
	case len(args) > max:
		return errTooMany
	default:
		return nil
	}
}

// ValidateCmd validate the number of arguments on a cobra command
func ValidateCmd(argsMin, argsMax int) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return Validate(args, argsMin, argsMax)
	}
}
