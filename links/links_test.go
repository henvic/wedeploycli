package links

import "testing"

type testcase struct {
	Func func() string
	we   string
	dxp  string
}

var cases = []testcase{
	testcase{
		Configuring,
		"https://wedeploy.com/docs/deploy/configuring-deployments/",
		"https://help.liferay.com/hc/en-us/articles/360012918551-Configuring-via-the-wedeploy-json",
	},
	testcase{
		Docs,
		"https://wedeploy.com/docs/",
		"https://help.liferay.com/hc/en-us/categories/360000813091-Liferay-DXP-Cloud-Documentation",
	},
	testcase{
		UsingCLI,
		"https://wedeploy.com/docs/intro/using-the-command-line/",
		"https://help.liferay.com/hc/en-us/articles/360015214691-Command-line-Tool",
	},
	testcase{
		Support,
		"support@wedeploy.com",
		"https://help.liferay.com/hc/en-us/categories/360000813091-Liferay-DXP-Cloud-Documentation",
	},
	testcase{
		PlanUpgrade,
		"https://console.wedeploy.com/account/billing",
		"https://help.liferay.com/hc/en-us/categories/360000813091-Liferay-DXP-Cloud-Documentation",
	},
}

func TestLink(t *testing.T) {
	for _, c := range cases {
		got := c.Func()

		if got != c.we {
			t.Errorf("Wanted %v, got %v instead", c.we, got)
		}
	}

	SetDXP()

	for _, c := range cases {
		got := c.Func()

		if got != c.dxp {
			t.Errorf("Wanted %v, got %v instead", c.dxp, got)
		}
	}

	SetWeDeploy()

	for _, c := range cases {
		got := c.Func()

		if got != c.we {
			t.Errorf("Wanted %v, got %v instead", c.we, got)
		}
	}
}
