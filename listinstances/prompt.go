package listinstances

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/henvic/wedeploycli/color"
	"github.com/henvic/wedeploycli/config"
	"github.com/henvic/wedeploycli/fancy"
	"github.com/henvic/wedeploycli/verbose"
)

// Prompt from the list selection
func (l *List) Prompt(ctx context.Context, wectx config.Context) (string, error) {
	return l.printAndPrompt(ctx, wectx, false)
}

// AutoSelectOrPrompt lets you select an instance from the list selection
// or select a single instance automatically if only one exists.
func (l *List) AutoSelectOrPrompt(ctx context.Context, wectx config.Context) (string, error) {
	return l.printAndPrompt(ctx, wectx, true)
}

func (l *List) printAndPrompt(ctx context.Context, wectx config.Context, askSingle bool) (string, error) {
	l.once = true
	l.PoolingInterval = time.Minute
	l.prepare(ctx, wectx)
	l.updateHandler()

	if askSingle && len(l.Instances) == 1 && l.Instances[0].State == "running" {
		container := l.Instances[0].ContainerID
		verbose.Debug("Only one instance found. Automatically connection to instance " + container)
		return container, nil
	}

	fmt.Printf("Please %s an instance from the list below.\n",
		color.Format(color.FgMagenta, color.Bold, "select"))

	l.SelectNumber = true

	if err := l.printOnce(ctx, wectx); err != nil {
		return "", err
	}

	return l.prompt()
}

func (l *List) printOnce(ctx context.Context, wectx config.Context) error {
	l.w.Init(l.outStream)
	l.watchHandler()

	l.watchMutex.RLock()
	var le = l.lastError
	l.watchMutex.RUnlock()
	return le
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
		if !strings.HasPrefix(s, option) {
			continue
		}

		if candidate != "" {
			ambiguity(candidate, option)
			return ""
		}

		candidate = s
	}

	if candidate != "" {
		option = candidate
	}

	return option
}

func ambiguity(candidate, option string) {
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
}
