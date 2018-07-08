package listinstances

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/fancy"
)

// Prompt from the list selection
func (l *List) Prompt(ctx context.Context, wectx config.Context) (string, error) {
	fmt.Printf("Please %s a instance from the list below.\n",
		color.Format(color.FgMagenta, color.Bold, "select"))

	l.SelectNumber = true

	if err := l.Once(ctx, wectx); err != nil {
		return "", err
	}

	return l.prompt()
}

func (l *List) prompt() (string, error) {
	fmt.Println("")
typeInstanceID:
	fmt.Println(fancy.Question("Type instance ID or #") + " " + fancy.Tip("default: #1"))

	var option, err = fancy.Prompt()

	if err != nil {
		return "", err
	}

	if option == "" {
		if len(l.selectors) != 0 {
			return l.selectors[0], nil
		}

		return "", errors.New("instance #1 can't be found")
	}

	var sel = l.selectors
	num, err := strconv.Atoi(option)

	if err == nil && num > 0 && num <= len(sel) {
		return sel[num-1], nil
	}

	instance := maybeGetOption(option, l.selectors)

	if instance == "" {
		goto typeInstanceID
	}

	return instance, nil
}

func maybeGetOption(option string, selectors []string) string {
	var candidate string

	for _, s := range selectors {
		if strings.HasPrefix(s, option) {
			if candidate != "" {
				example := candidate

				if len(example) >= 8 {
					example = example[:8]
				}

				_, _ = fmt.Fprintf(os.Stderr, "%s%s%s%s%s%s\n",
					color.Format(color.FgHiBlack, "Multiple instances match ID \""),
					option,
					color.Format(color.FgHiBlack, "\". Try again. Example: use \""),
					option,
					example,
					color.Format(color.FgHiBlack, "\" or #."))
				return ""
			}

			candidate = s
		}
	}

	if candidate != "" {
		option = candidate
	}

	return option
}
