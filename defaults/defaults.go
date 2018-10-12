package defaults

// don't change this to const as it would make go build -ldflags fail silently
var (
	// Version of the WeDeploy Project CLI tool
	Version = "master"

	// Build commit
	Build = ""

	// BuildTime is the time when the build was generated
	BuildTime = ""

	// Infrastructure is the default remote address
	Infrastructure = "wedeploy.com"

	// DashboardAddressPrefix for a given remote
	DashboardAddressPrefix = "console."

	// DashboardURLPrefix for a given remote
	DashboardURLPrefix = "https://" + DashboardAddressPrefix

	// PlanUpgradeURL to show or open (mostly due to the exceededPlanMaximum error)
	PlanUpgradeURL = "https://console.wedeploy.com/account/billing"

	// DiagnosticsEndpoint is the endpoint to where the diagnostics should be sent
	DiagnosticsEndpoint = "https://cli-metrics.wedeploy.com/diagnostics/report"

	// Docs page
	Docs = "http://wedeploy.com/docs/"

	// Hub for the system
	Hub = "http://api.wedeploy.io"

	// AnalyticsEndpoint for posting analytics events in bulk
	AnalyticsEndpoint = "https://cli-metrics.wedeploy.com/metrics/bulk"

	// CloudRemote is the name for the default cloud for WeDeploy
	CloudRemote = "wedeploy"

	// SupportEmail value
	SupportEmail = "support@wedeploy.com"

	// StableReleaseChannel for the distribution of the CLI tool
	StableReleaseChannel = "stable"
)

// MousetrapHelpText is used by cobra to show an error message when the user tries to use the CLI
// without a terminal open (i.e., double-clicking on Windows Explorer)
const MousetrapHelpText = `This is a command line tool.

You need to open this using a terminal/console application.

If you want to learn how to use the CLI for WeDeploy, please see:
https://wedeploy.com/docs/intro/using-the-command-line/`
