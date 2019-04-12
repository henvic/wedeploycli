package defaults

// don't change this to const as it would make go build -ldflags fail silently
var (
	// Version of the CLI Tool (replaced with -ldflags during sofware release builds)
	Version = "master"

	// Build commit
	Build = ""

	// BuildTime is the time when the build was generated
	BuildTime = ""

	// Infrastructure is the default remote address
	Infrastructure = "us-west-1.liferay.cloud"

	// DashboardAddressPrefix for a given remote
	DashboardAddressPrefix = "console."

	// DashboardURLPrefix for a given remote
	DashboardURLPrefix = "https://" + DashboardAddressPrefix

	// DiagnosticsEndpoint is the endpoint to where the diagnostics should be sent
	DiagnosticsEndpoint = "https://cli-metrics.wedeploy.com/diagnostics/report"

	// AnalyticsEndpoint for posting analytics events in bulk
	AnalyticsEndpoint = "https://cli-metrics.wedeploy.com/metrics/bulk"

	// CloudRemote is the name for the default cloud
	CloudRemote = "liferay"

	// StableReleaseChannel for the distribution of the CLI Tool
	StableReleaseChannel = "stable"
)

// MousetrapHelpText is used by cobra to show an error message when the user tries to use the CLI
// without a terminal open (i.e., double-clicking on Windows Explorer)
const MousetrapHelpText = `This is a command line tool.

You need to open this using a terminal/console application.

If you want to learn how to use the CLI for Liferay, please see:
https://help.liferay.com/hc/en-us/articles/360015214691-Command-line-Tool`
