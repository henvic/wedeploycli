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

	// DiagnosticsEndpoint is the endpoint to where the diagnostics should be sent
	DiagnosticsEndpoint = "https://diagnostics.wedeploy.com/report/cli"

	// Docs page
	Docs = "http://wedeploy.com/docs/"

	// Hub for the system
	Hub = "http://api.wedeploy.io"

	// WeDeployImageTag is the WeDeploy image tag for docker
	WeDeployImageTag = "beta"

	// AnalyticsEndpoint for posting analytics events in bulk
	AnalyticsEndpoint = "https://cli-metrics.wedeploy.com/"

	// CloudRemote is the name for the default cloud for WeDeploy
	CloudRemote = "wedeploy"

	// SupportEmail value
	SupportEmail = "support@wedeploy.com"

	// StableReleaseChannel for the distribution of the CLI tool
	StableReleaseChannel = "stable"
)
