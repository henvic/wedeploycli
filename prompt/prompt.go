package prompt

import (
	"fmt"
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

	fmt.Fprintf(os.Stdout, param+": ")
	fmt.Fscanf(os.Stdin, "%s\n", &value)
	return value
}
