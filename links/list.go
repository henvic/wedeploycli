package links

import (
	"fmt"
	"text/template"

	"github.com/wedeploy/cli/templates"
)

func init() {
	templates.Funcs(template.FuncMap{
		"LinkConfiguring": Configuring,
		"LinkDocs":        Docs,
		"LinkUsingCLI":    UsingCLI,
		"LinkSupport":     Support,
		"LinkPlanUpgrade": PlanUpgrade,
	})
}

// Configuring deployments link.
func Configuring() string {
	return fmt.Sprint(Link{
		WeDeploy: "https://wedeploy.com/docs/deploy/configuring-deployments/",
		DXP:      "https://help.liferay.com/hc/en-us/articles/360012918551-Configuring-via-the-wedeploy-json",
	})
}

// Docs of the cloud infrastructure.
func Docs() string {
	return fmt.Sprint(Link{
		WeDeploy: "https://wedeploy.com/docs/",
		DXP:      "https://help.liferay.com/hc/en-us/categories/360000813091-Liferay-DXP-Cloud-Documentation",
	})
}

// UsingCLI link.
func UsingCLI() string {
	return fmt.Sprint(Link{
		WeDeploy: "https://wedeploy.com/docs/intro/using-the-command-line/",
		DXP:      "https://help.liferay.com/hc/en-us/articles/360015214691-Command-line-Tool",
	})
}

// Support contact for the service.
func Support() string {
	return fmt.Sprint(Link{
		WeDeploy: "support@wedeploy.com",
		DXP:      "https://help.liferay.com/hc/en-us/categories/360000813091-Liferay-DXP-Cloud-Documentation",
	})
}

// PlanUpgrade to show or open (mostly due to the exceededPlanMaximum error)
func PlanUpgrade() string {
	return fmt.Sprint(Link{
		WeDeploy: "https://console.wedeploy.com/account/billing",
		DXP:      "https://help.liferay.com/hc/en-us/categories/360000813091-Liferay-DXP-Cloud-Documentation",
	})
}
