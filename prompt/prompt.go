package prompt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/hashicorp/errwrap"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	inStream  io.Reader = os.Stdin
	outStream io.Writer = os.Stdout
	errStream io.Writer = os.Stderr

	isTerminal = terminal.IsTerminal(int(os.Stdin.Fd()))
)

// SelectOption prompts for an option from a list
func SelectOption(indexLength int, equivalents map[string]int) (index int, err error) {
	if indexLength == 0 {
		return -1, errors.New("No options available.")
	}

	var option string
	option, err = Prompt(fmt.Sprintf("\nSelect from 1..%d", indexLength))

	if err != nil {
		return -1, err
	}

	option = strings.TrimSpace(option)

	if equivalents != nil {
		if index, ok := equivalents[option]; ok {
			return getSelectOptionIndex(index, indexLength, nil)
		}
	}

	index, err = strconv.Atoi(option)
	return getSelectOptionIndex(index, indexLength, err)
}

func getSelectOptionIndex(index, indexLength int, err error) (int, error) {
	index--
	if err != nil || index < 0 || index > indexLength {
		return -1, errors.New("Invalid option.")
	}

	return index, nil
}

// Prompt returns a prompt to receive the value of a parameter.
// If the key is on a secret keys list it suppresses the feedback.
func Prompt(param string) (string, error) {
	if !isTerminal {
		return "", errors.New("Input device is not a terminal. " +
			`Can't read "` + param + `"`)
	}

	fmt.Fprintf(outStream, param+": ")
	reader := bufio.NewReader(inStream)
	value, err := reader.ReadString('\n')

	if err != nil {
		return "", errwrap.Wrapf("Can't read stdin for "+param+": {{err}}", err)
	}

	return value[:len(value)-1], nil
}

// Hidden provides a prompt without echoing the value entered
func Hidden(param string) (string, error) {
	if !isTerminal {
		return "", errors.New("Input device is not a terminal. " +
			`Can't read "` + param + `"`)
	}

	fmt.Fprintf(outStream, param+": ")
	var b, err = terminal.ReadPassword(syscall.Stdin)

	if err != nil {
		return "", errwrap.Wrapf("Can't read stdin for "+param+": {{err}}", err)
	}

	return string(b), nil
}
