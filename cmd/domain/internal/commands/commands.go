package commands

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/services"
)

// Command for domains
type Command struct {
	SetupHost      cmdflagsfromhost.SetupHost
	ServicesClient *services.Client

	Domains []string
}

// Show domains
func (c *Command) Show(ctx context.Context) error {
	c.Domains = []string{}

	var l = list.New(list.Filter{
		Project:  c.SetupHost.Project(),
		Services: []string{c.SetupHost.Service()},
	})

	if err := l.Once(ctx, we.Context()); err != nil {
		return err
	}

	var service, err = c.ServicesClient.Get(ctx,
		c.SetupHost.Project(),
		c.SetupHost.Service())

	if err != nil {
		return err
	}

	c.Domains = service.CustomDomains

	if len(c.Domains) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "No custom domains found.\n")
		return nil
	}

	sort.Slice(c.Domains, func(i, j int) bool {
		return c.Domains[i] < c.Domains[j]
	})

	fmt.Printf("%s\t%s\n",
		color.Format(color.FgHiBlack, "#"),
		color.Format(color.FgHiBlack, "Domain"))

	for pos, v := range c.Domains {
		fmt.Printf("%d\t%s\n", pos+1, v)
	}

	return nil
}

// Add domains
func (c *Command) Add(ctx context.Context, args []string) error {
	var domains, err = c.getAddDomains(args)

	if err != nil {
		return err
	}

	for _, domain := range domains {
		err = c.ServicesClient.AddDomain(ctx, c.SetupHost.Project(), c.SetupHost.Service(), domain)

		if err != nil {
			return errwrap.Wrapf("can't add \""+domain+"\": {{err}}", err)
		}

		fmt.Printf("Custom domain \"%v\" added.\n", domain)
	}

	return nil
}

// Delete domains
func (c *Command) Delete(ctx context.Context, args []string) error {
	var domains, err = c.getDeleteDomains(args)

	if err != nil {
		return err
	}

	for _, domain := range domains {
		err := c.ServicesClient.RemoveDomain(ctx, c.SetupHost.Project(), c.SetupHost.Service(), domain)

		if err != nil {
			return err
		}

		fmt.Printf("Custom domain \"%s\" deleted.\n", domain)
	}

	return nil
}

func (c *Command) getAddDomains(args []string) ([]string, error) {
	if len(args) != 0 {
		return args, nil
	}

	fmt.Println(fancy.Question("Type custom domains for \"" + c.SetupHost.Host() + "\" (e.g., example.com example.net)"))
	var domainss, err = fancy.Prompt()

	if err != nil {
		return []string{}, err
	}

	var domains = strings.Split(domainss, " ")
	return filterEmptyDomains(domains), nil
}

func (c *Command) getDeleteDomains(args []string) ([]string, error) {
	if len(args) != 0 {
		return args, nil
	}

	fmt.Println(fancy.Question("Select a custom domain # or address to delete from \"" + c.SetupHost.Host() + "\""))
	var domainss, err = fancy.Prompt()

	if err != nil {
		return []string{}, err
	}

	var domains = strings.Split(domainss, " ")

	for index, domain := range domains {
		d := c.getDomainOrOption(domain)

		if d != domain {
			domains[index] = d
		}
	}

	return filterEmptyDomains(domains), nil
}

func (c *Command) getDomainOrOption(answer string) string {
	for _, d := range c.Domains {
		if answer == d {
			return d
		}
	}

	switch num, err := strconv.Atoi(answer); {
	case err != nil || num < 1 || num > len(c.Domains):
		return answer
	default:
		return c.Domains[num-1]
	}
}

func filterEmptyDomains(domains []string) []string {
	var filtered []string

	for _, d := range domains {
		if d != "" {
			filtered = append(filtered, d)
		}
	}

	return filtered
}
