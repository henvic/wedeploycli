package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/isterm"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/services"
)

// ErrNoEnvToAdd is used when there is no environment varaible to add
var ErrNoEnvToAdd = errors.New("no environment variable to add")

// Command for environment variables
type Command struct {
	SetupHost      cmdflagsfromhost.SetupHost
	ServicesClient *services.Client

	Envs []services.EnvironmentVariable

	SkipPrompt bool
}

func has(filterEnvKeys []string, key string) bool {
	if len(filterEnvKeys) == 0 {
		return true
	}

	for _, f := range filterEnvKeys {
		if f == key {
			return true
		}
	}

	return false
}

// Show environment variables
func (c *Command) Show(ctx context.Context, filterEnvKeys ...string) error {
	c.Envs = []services.EnvironmentVariable{}

	var l = list.New(list.Filter{
		Project:  c.SetupHost.Project(),
		Services: []string{c.SetupHost.Service()},
	})

	if err := l.Once(ctx, we.Context()); err != nil {
		return err
	}

	var envs, err = c.ServicesClient.GetEnvironmentVariables(ctx,
		c.SetupHost.Project(),
		c.SetupHost.Service())

	if err != nil {
		return err
	}

	for _, e := range envs {
		if !has(filterEnvKeys, e.Name) {
			continue
		}

		c.Envs = append(c.Envs, services.EnvironmentVariable{
			Name:  e.Name,
			Value: e.Value,
		})
	}

	if len(c.Envs) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "No environment variable found.\n")
		return nil
	}

	sort.Slice(c.Envs, func(i, j int) bool {
		return c.Envs[i].Name < c.Envs[j].Name
	})

	fmt.Printf("%s\t%s\t%s\n",
		color.Format(color.FgHiBlack, "#"),
		color.Format(color.FgHiBlack, "Key"),
		color.Format(color.FgHiBlack, "Value"))

	for c, v := range c.Envs {
		fmt.Printf("%d\t%s\t%s\n", c+1, v.Name, v.Value)
	}

	return nil
}

// Add environment variables
func (c *Command) Add(ctx context.Context, args []string) error {
	var envs, err = c.getAddEnvs(args)

	if err != nil {
		return err
	}

	for _, env := range envs {
		err = c.ServicesClient.SetEnvironmentVariable(ctx, c.SetupHost.Project(), c.SetupHost.Service(), env.Name, env.Value)

		if err != nil {
			return errwrap.Wrapf("can't add \""+env.Name+"\": {{err}}", err)
		}

		fmt.Printf("Environment variable \"%v\" added.\n", env.Name)
	}

	return nil
}

// Replace environment variables
func (c *Command) Replace(ctx context.Context, args []string) error {
	var envs, err = c.getAddEnvs(args)

	if err != nil && err != ErrNoEnvToAdd {
		return err
	}

	if err = c.ServicesClient.SetEnvironmentVariables(ctx, c.SetupHost.Project(), c.SetupHost.Service(), envs); err != nil {
		return err
	}

	for _, env := range envs {
		fmt.Printf("Environment variable \"%v\" added.\n", env.Name)
	}

	return nil
}

func (c *Command) getAddEnvs(args []string) (envs []services.EnvironmentVariable, err error) {
	args = filterEmptyEnvValues(args)

	if len(args) == 0 && !c.SkipPrompt && isterm.Check() {
		fmt.Println(fancy.Question("Type environment variables for \"" + c.SetupHost.Host() + "\" (e.g., A=1 B=2 C=3)"))
		var argss string
		argss, err = fancy.Prompt()

		if err != nil {
			return nil, err
		}

		args = strings.Split(argss, " ")
	}

	if strings.Join(args, "") == "" {
		return envs, ErrNoEnvToAdd
	}

	return splitEnvKeyValueParameters(args)
}

func splitEnvKeyValueParameters(args []string) (envs []services.EnvironmentVariable, err error) {
	if len(args) == 2 && !strings.Contains(args[0], "=") && !strings.Contains(args[1], "=") {
		envs = append(envs, services.EnvironmentVariable{
			Name:  args[0],
			Value: args[1],
		})

		return envs, nil
	}

	for _, v := range args {
		var kv = strings.SplitN(v, "=", 2)

		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid environment variable key/value pair: \"%s\"", v)
		}

		envs = append(envs, services.EnvironmentVariable{
			Name:  kv[0],
			Value: kv[1],
		})
	}

	return envs, nil
}

// Delete environment variables
func (c *Command) Delete(ctx context.Context, args []string) error {
	var envs, err = c.getDeleteEnvKeys(args)

	if err != nil {
		return err
	}

	for _, env := range envs {
		err := c.ServicesClient.UnsetEnvironmentVariable(ctx, c.SetupHost.Project(), c.SetupHost.Service(), env)

		if err != nil {
			return err
		}

		fmt.Printf("Environment variable \"%s\" deleted.\n", env)
	}

	return nil
}

func (c *Command) getDeleteEnvKeys(args []string) ([]string, error) {
	if len(args) != 0 {
		return args, nil
	}

	fmt.Println(fancy.Question("Select a environment variable # or name to delete from \"" + c.SetupHost.Host() + "\""))
	var envss, err = fancy.Prompt()

	if err != nil {
		return []string{}, err
	}

	var envs = strings.Split(envss, " ")

	for index, env := range envs {
		e := c.getEnvKeyOrOption(env)

		if e != env {
			envs[index] = e
		}
	}

	return filterEmptyEnvValues(envs), nil
}

func (c *Command) getEnvKeyOrOption(answer string) string {
	for _, e := range c.Envs {
		if answer == e.Name {
			return e.Name
		}
	}

	switch num, err := strconv.Atoi(answer); {
	case err != nil || num < 1 || num > len(c.Envs):
		return answer
	default:
		return c.Envs[num-1].Name
	}
}

func filterEmptyEnvValues(envKeys []string) []string {
	var filtered []string

	for _, e := range envKeys {
		if e != "" {
			filtered = append(filtered, e)
		}
	}

	return filtered
}

// EnvIsDeprecatedWarning is used to print a warning when the deprecated "we env" command is used
func EnvIsDeprecatedWarning(cmd *cobra.Command, args []string) {
	if os.Args[1:][0] == "env" {
		_, _ = fmt.Fprintln(os.Stderr, color.Format(color.FgHiRed,
			`"we env" is deprecated: use "%s" next time.`, cmd.UseLine()))
	}
}
