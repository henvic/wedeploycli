package template

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/formatter"
)

// Configure template for cobra commands
func Configure(rootCmd *cobra.Command) {
	cobra.AddTemplateFunc("printCommandsAndFlags", printCommandsAndFlags)
	rootCmd.SetUsageTemplate(`{{printCommandsAndFlags .UseLine .Example .Commands .Flags}}`)
	rootCmd.SetHelpTemplate(`{{if not (eq .CommandPath "` + rootCmd.Name() + `")}}{{with or .Long .Short }}{{color FgYellow BgHiYellow "!"}} {{. | trim | color FgHiYellow}}
{{end}}{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`)
}

type usagePrinter struct {
	useLine             string
	example             string
	cs                  []*cobra.Command
	f                   *pflag.FlagSet
	buf                 *bytes.Buffer
	tw                  *formatter.TabWriter
	flags               flagsDescriptions
	showFlagsParamField bool
}

type flagsDescriptions []flagDescription

type flagDescription struct {
	Name        string
	Description []byte
	Used        bool
}

func colorSpacingOffset() string {
	if formatter.Human {
		return "     "
	}

	return ""
}

func printCommandsAndFlags(useLine, example string, cs []*cobra.Command, f *pflag.FlagSet) string {
	useLine = strings.TrimSuffix(useLine, " [flags]")

	up := &usagePrinter{
		useLine: useLine,
		example: example,
		cs:      cs,
		f:       f,
	}
	up.buf = new(bytes.Buffer)
	up.tw = formatter.NewTabWriter(up.buf)
	up.printCommands()
	up.printFlags()
	_ = up.tw.Flush()
	return up.buf.String()
}

func (up *usagePrinter) printCommands() {
	var cmdPart = " [command]"

	if len(up.cs) == 0 {
		cmdPart = ""
	}

	fmt.Fprintf(up.buf, "%s%s",
		color.Format(color.FgYellow, color.BgHiYellow, "!"),
		fmt.Sprintf("%v\n", color.Format(color.FgHiYellow,
			fmt.Sprintf(" Usage: %s%s [flag]", up.useLine, cmdPart))))

	if up.example != "" {
		fmt.Fprintf(up.buf,
			"%s%s\n\n",
			color.Format(color.FgHiYellow, "  Examples:\n"),
			up.example,
		)
	}

	if len(up.cs) == 0 {
		return
	}

	fmt.Fprint(up.tw, color.Format(color.FgHiBlack, "  Command\t"+colorSpacingOffset()+"Description")+"\n")

	for _, c := range up.cs {
		if c.IsAvailableCommand() {
			fmt.Fprintf(up.tw, "  %v\t%v\n", c.Name(), c.Short)
		}
	}

	fmt.Fprintln(up.tw, "\t") // \t here keeps the alignment between commands and flags
}

func (up *usagePrinter) printFlags() {
	up.f.VisitAll(func(flag *pflag.Flag) {
		if !flag.Hidden && flag.Value.Type() != "bool" {
			up.showFlagsParamField = true
		}
	})

	if up.showFlagsParamField {
		fmt.Fprintf(up.tw, "%s\n",
			color.Format(color.FgHiBlack,
				"  Flag\t"+colorSpacingOffset()+"Parameter\t"+colorSpacingOffset()+"Description"))
	} else {
		fmt.Fprintf(up.tw, "%s\n",
			color.Format(color.FgHiBlack,
				"  Flag\t"+colorSpacingOffset()+"Description"))
	}

	up.f.VisitAll(up.preparePrintFlag)

	var begin = up.useFlagsHelpDescriptionFiltered([]string{
		"service",
		"project",
		"remote",
		"url",
	})

	var end = up.useFlagsHelpDescriptionFiltered([]string{
		"help",
		"quiet",
		"verbose",
	})

	var middle = up.useFlagsHelpDescription()
	fmt.Fprintf(up.tw, "%s%s%s", string(begin), string(middle), string(end))
}

func (up *usagePrinter) useFlagsHelpDescriptionFiltered(list []string) []byte {
	var buf bytes.Buffer

	for _, filtered := range list {
		for i, flag := range up.flags {
			if flag.Name == filtered && !flag.Used {
				up.flags[i].Used = true
				buf.Write(flag.Description)
			}
		}
	}

	return buf.Bytes()
}

func (up *usagePrinter) useFlagsHelpDescription() []byte {
	var buf bytes.Buffer

	for i, flag := range up.flags {
		if !flag.Used {
			up.flags[i].Used = true
			buf.Write(flag.Description)
		}
	}

	return buf.Bytes()
}

func (up *usagePrinter) preparePrintFlag(flag *pflag.Flag) {
	if flag.Deprecated != "" || flag.Hidden {
		return
	}

	var buf = bytes.NewBufferString("  ")

	if flag.Shorthand != "" && flag.ShorthandDeprecated == "" {
		buf.WriteString(fmt.Sprintf("-%s, ", flag.Shorthand))
	} else {
		buf.WriteString("    ")
	}

	buf.WriteString(fmt.Sprintf("--%s", flag.Name))

	if flag.NoOptDefVal != "" {
		switch flag.Value.Type() {
		case "string":
			buf.WriteString(fmt.Sprintf("[=\"%s\"]", flag.NoOptDefVal))
		case "bool":
			if flag.NoOptDefVal != "true" {
				buf.WriteString(fmt.Sprintf("[=%s]", flag.NoOptDefVal))
			}
		default:
			buf.WriteString(fmt.Sprintf("[=%s]", flag.NoOptDefVal))
		}
	}

	var flagType = flag.Value.Type()

	if flagType == "bool" {
		flagType = ""
	}

	if up.showFlagsParamField {
		buf.WriteString(fmt.Sprintf("\t%s\t%s", flagType, flag.Usage))
	} else {
		buf.WriteString(fmt.Sprintf("\t%s", flag.Usage))
	}

	if !isDefaultFlagValueZero(flag) {
		if flag.Value.Type() == "string" {
			buf.WriteString(fmt.Sprintf(" (default %q)", flag.DefValue))
		} else {
			buf.WriteString(fmt.Sprintf(" (default %s)", flag.DefValue))
		}
	}

	buf.WriteString("\n")
	up.flags = append(up.flags, flagDescription{
		Name:        flag.Name,
		Description: buf.Bytes(),
	})
}

func isDefaultFlagValueZero(f *pflag.Flag) bool {
	switch f.DefValue {
	case "", "0", "0s", "false", "[]":
		return true
	}

	return false
}
