package execargs

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// MaybeRewrite cobra arguments.
func MaybeRewrite(cmd *cobra.Command, args []string) ([]string, bool) {
	pos, rewrite := getPosition(cmd, args)

	if !rewrite {
		return []string{}, false
	}

	stringFlags := getStringFlags(cmd.Flags())
	skip := false

	found := -1

	for index, a := range args[pos:] {
		if a == "--" {
			return []string{}, false
		}

		if skip {
			skip = false
			continue
		}

		if strings.HasPrefix(a, "-") {
			// remember flags might be --foo, --foo=value, and --foo value.
			if stringFlags[a] {
				skip = true
			}

			continue
		}

		found = index
		break
	}

	if found == -1 {
		return []string{}, false
	}

	na := append(args[:found+pos],
		append([]string{"--"}, args[found+pos:]...)...,
	)

	return na, true
}

func getPosition(cmd *cobra.Command, args []string) (int, bool) {
	name := cmd.Name()

	stringFlags := getStringFlags(cmd.Flags())
	skip := false

	for index, a := range args {
		if a == "--" {
			break
		}

		if skip {
			skip = false
			continue
		}

		if strings.HasPrefix(a, "-") {
			// remember flags might be --foo, --foo=value, and --foo value.
			if stringFlags[a] {
				skip = true
			}

			continue
		}

		if a == name {
			return index + 1, true
		}

		break
	}

	return -1, false
}

func getStringFlags(all *pflag.FlagSet) map[string]bool {
	var flags = map[string]bool{}

	all.VisitAll(func(f *pflag.Flag) {
		if f.Value.Type() != "string" {
			return
		}

		if f.Name != "" {
			flags["--"+f.Name] = true
		}

		if f.Deprecated != "" {
			flags["--"+f.Deprecated] = true
		}

		if f.Shorthand != "" {
			flags["-"+f.Shorthand] = true
		}

		if f.ShorthandDeprecated != "" {
			flags["-"+f.ShorthandDeprecated] = true
		}
	})

	return flags
}
