package cmd

func init() {
	RootCmd.SetUsageTemplate(`● {{if .Runnable}}{{appendIfNotPresent .UseLine "[command]"}}{{else}}{{.UseLine}}{{end}} [flags]{{if gt .Aliases 0}}

Aliases: {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{ .Example }}{{end}}{{ if .HasAvailableSubCommands}}

Commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasAvailableFlags}}{{if ne .Name "we"}}
  {{end}}
Flags:
{{.Flags.FlagUsages | trimRightSpace}}
{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsHelpCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}
{{end}}{{ if .HasAvailableSubCommands }}
Use "{{.CommandPath}} [command] --help" for more information about a command.
{{end}}`)
}
