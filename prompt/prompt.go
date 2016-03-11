package prompt

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/bgentry/speakeasy"
)

var secretKeys = []string{
	"password",
	"token",
	"secret",
}

var (
	inStream  io.Reader = os.Stdin
	outStream io.Writer = os.Stdout
	errStream io.Writer = os.Stderr
)

func isSecretKey(key string) bool {
	var match, _ = regexp.MatchString(
		"("+strings.Join(secretKeys, "|")+")",
		strings.ToLower(key))
	return match
}

// Prompt returns a prompt to receive the value of a parameter.
// If the key is on a secret keys list it suppresses the feedback.
func Prompt(param string) string {
	var value string

	if isSecretKey(param) {
		value, err := speakeasy.Ask(param + ": ")

		if err != nil {
			panic(err)
		}

		return value
	}

	fmt.Fprintf(outStream, param+": ")
	fmt.Fscanf(inStream, "%s\n", &value)
	return value
}
