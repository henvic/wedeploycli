package prompt

import (
	"bufio"
	"errors"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/wedeploycli/isterm"
	"golang.org/x/crypto/ssh/terminal"
)

var inStream io.Reader = os.Stdin
var errInvalidOption = errors.New("invalid option")

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
		if i, ok := equivalents[option]; ok {
			return getSelectOptionIndex(i, indexLength, nil)
		}
	}

	index, err = strconv.Atoi(option)
	return getSelectOptionIndex(index, indexLength, err)
}

func getSelectOptionIndex(index, indexLength int, err error) (int, error) {
	index--
	if err != nil || index < 0 || index >= indexLength {
		return -1, errInvalidOption
	}

	return index, nil
}

// Prompt returns a prompt to receive the value of a parameter.
// If the key is on a secret keys list it suppresses the feedback.
func Prompt() (string, error) {
	// Checking if is terminal and not Windows because Windows is Windows...
	if !isterm.Check() && runtime.GOOS != "windows" {
		return "", errors.New("input device is not a terminal")
	}

	reader := bufio.NewReader(inStream)
	value, err := reader.ReadString('\n')

	// two cases:
	// on Unix: \n
	// on Windows: \r\n
	// remove line break
	if runtime.GOOS == "windows" {
		value = strings.TrimRight(value, "\r\n")
	} else {
		value = strings.TrimRight(value, "\n")
	}

	if err != nil {
		return "", errwrap.Wrapf("can't read stdin : {{err}}", err)
	}

	return value, nil
}

// Hidden provides a prompt without echoing the value entered
func Hidden() (string, error) {
	// Checking if is terminal and not Windows because Windows is Windows...
	// Actually terminal.ReadPassword is even broken on Windows Subsystem for Linux (Windows 10 and 2016 Server)
	if !isterm.Check() && runtime.GOOS != "windows" {
		return "", errors.New("input device is not a terminal: can't read password")
	}

	var b, err = terminal.ReadPassword(int(syscall.Stdin))

	if err != nil {
		return "", errwrap.Wrapf("can't read stdin (hidden): {{err}}", err)
	}

	return string(b), nil
}
