package fancy

import (
	"bytes"
	"fmt"
	"os"

	"strings"

	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/prompt"
)

// Question formatter
func Question(a interface{}) string {
	var q = color.Format(color.FgMagenta, "?") + " "
	var parts = strings.Split(fmt.Sprintf("%v", a), "\n")
	var buf = &bytes.Buffer{}
	for i, p := range parts {
		fmt.Fprint(buf, q+color.Format(color.Reset, p))

		if i < len(parts)-1 {
			fmt.Fprintln(buf, "")
		}
	}

	return buf.String()
}

// Info formatter
func Info(a interface{}) string {
	return color.Format(color.FgHiBlack, a)
}

// Success formatter
func Success(a interface{}) string {
	return color.Format(color.FgHiGreen, a)
}

// Error formatter
func Error(a interface{}) string {
	return color.Format(color.FgHiBlack, formatError(a))
}

// Tip formatter
func Tip(a interface{}) string {
	return fmt.Sprintf("%v%v%v", color.Format(color.FgHiMagenta, "["), a, color.Format(color.FgHiMagenta, "]"))
}

// Prompt with fancy "> "
func Prompt() (string, error) {
	fmt.Print(color.Format(color.FgHiBlack, "> "))
	var res, err = prompt.Prompt()

	if err != nil {
		fmt.Println("")
	}

	return res, err
}

// HiddenPrompt with a >
func HiddenPrompt() (string, error) {
	fmt.Print(color.Format(color.FgHiBlack, "> "))
	var res, err = prompt.Hidden()

	if err != nil {
		fmt.Println("")
	}
	return res, err
}

// Boolean question
func Boolean(question string) (yes bool, err error) {
	question = Question(question)
	fmt.Printf("%s %s\n", question, color.Format(color.FgMagenta, "[y/n]"))

	for {
		var choice, err = Prompt()

		if err != nil {
			return false, err
		}

		cInput := strings.TrimSpace(strings.ToLower(choice))

		switch cInput {
		case "y", "yes", "yep", "yeh", "yeah":
			return true, nil
		case "n", "no", "nah", "nope":
			return false, nil
		case "":
			fmt.Fprintln(os.Stderr, Error("Select an option."))
		default:
			fmt.Fprintln(os.Stderr,
				Error(`No valid answer was found for "`+
					color.Escape(choice)+
					`"`))
		}
	}
}

// Options selector
type Options struct {
	list []option
	Hash bool
}

type option struct {
	name        string
	description string
}

// Add option
func (o *Options) Add(name, description string) {
	o.list = append(o.list, option{name, description})
}

// List options
func (o *Options) List() string {
	var buf = &bytes.Buffer{}
	for _, option := range o.list {
		fmt.Fprintf(buf, "%s %s\n", color.Format(color.FgHiMagenta, color.Bold, strings.ToUpper(option.name)), option.description)
	}

	return buf.String()
}

// Ask for options printing a question
func (o *Options) Ask(q string) (string, error) {
	q = Question(q)
	names := []string{}

	for _, option := range o.list {
		names = append(names, option.name)
	}

	var printedList = "#"

	if !o.Hash {
		printedList = strings.Join(names, "/")
	}

	fmt.Printf("%s %s\n", q, color.Format(color.FgMagenta, "[%s]", printedList))

	fmt.Print(o.List())

	for {
		var choice, err = Prompt()

		if err != nil {
			return "", err
		}

		if res, ok := o.findMatch(choice); ok {
			return res, nil
		}

		switch len(choice) {
		case 0:
			fmt.Fprintln(os.Stderr, Error("Select an option."))
		default:
			fmt.Fprintln(os.Stderr,
				Error(`No valid answer was found for "`+
					color.Escape(choice)+
					`"`))
		}
	}
}

func (o *Options) findMatch(choice string) (string, bool) {
	choice = strings.ToLower(choice)
	for _, option := range o.list {
		if strings.ToLower(option.name) == choice {
			return option.name, true
		}
	}

	return "", false
}

func formatError(a interface{}) string {
	var errMsg = fmt.Sprintf("%v", a)

	switch len(errMsg) {
	case 0:
		return ""
	case 1:
		errMsg = strings.ToUpper(errMsg)
	default:
		errMsg = fmt.Sprintf("%s%s", strings.ToUpper(errMsg[0:1]), errMsg[1:])
	}

	if strings.Contains(errMsg, "\n") {
		return errMsg
	}

	switch l := errMsg[len(errMsg)-1:]; l {
	case "!", ".", "?":
	default:
		errMsg = errMsg + "."
	}

	return errMsg
}
