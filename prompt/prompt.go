package prompt

import (
	"bufio"
	"errors"
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

	isTerminal = terminal.IsTerminal(int(os.Stdin.Fd()))
)

// SelectOption prompts for an option from a list
func SelectOption(indexLength int, equivalents map[string]int) (index int, err error) {
	if indexLength == 0 {
		return -1, errors.New("no options available")
	}

	option, err := Prompt()

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
	if err != nil || index < 0 || index >= indexLength {
		return -1, errors.New("invalid option")
	}

	return index, nil
}

// Prompt returns a prompt to receive the value of a parameter.
// If the key is on a secret keys list it suppresses the feedback.
func Prompt() (string, error) {
	if !isTerminal {
		return "", errors.New("input device is not a terminal")
	}

	reader := bufio.NewReader(inStream)
	value, err := reader.ReadString('\n')

	// on Windows the line break is \r\n
	// let's trim an extra line (\n, not \r)
	value = strings.TrimRight(value, "\n")

	if err != nil {
		return "", errwrap.Wrapf("can't read stdin : {{err}}", err)
	}

	return value[:len(value)-1], nil
}

// Hidden provides a prompt without echoing the value entered
func Hidden() (string, error) {
	if !isTerminal {
		return "", errors.New("input device is not a terminal: can't read password")
	}

	var b, err = terminal.ReadPassword(int(syscall.Stdin))

	if err != nil {
		return "", errwrap.Wrapf("can't read stdin (hidden): {{err}}", err)
	}

	return string(b), nil
}
