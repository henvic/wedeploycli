package links

import (
	"text/template"

	"github.com/henvic/wedeploycli/templates"
)

func init() {
	templates.Funcs(template.FuncMap{
		"LinkConfiguring": link(Configuring),
		"LinkDocs":        link(Docs),
		"LinkUsingCLI":    link(UsingCLI),
		"LinkSupport":     link(Support),
		"LinkPlanUpgrade": link(PlanUpgrade),
	})
}

func link(s string) func() string {
	return func() string {
		return s
	}
}

const (
	// Configuring deployments link.
	Configuring = "https://help.liferay.com/hc/en-us/articles/360012918551-Configuring-via-the-wedeploy-json"

	// Docs of the cloud infrastructure.
	Docs = "https://help.liferay.com/hc/en-us/categories/360000813091-Liferay-DXP-Cloud-Documentation"

	// UsingCLI link.
	UsingCLI = "https://help.liferay.com/hc/en-us/articles/360015214691-Command-line-Tool"

	// Support link
	Support = "https://help.liferay.com/hc/en-us/categories/360000813091-Liferay-DXP-Cloud-Documentation"

	// PlanUpgrade link
	PlanUpgrade = "https://help.liferay.com/hc/en-us/categories/360000813091-Liferay-DXP-Cloud-Documentation"
)
