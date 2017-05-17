package defaults

var (
	// Version of the WeDeploy Project CLI tool
	Version = "master"

	// Build commit
	Build = ""

	// BuildTime is the time when the build was generated
	BuildTime = ""

	// DashboardAddressPrefix for a given remote
	DashboardAddressPrefix = "console."

	// DashboardURLPrefix for a given remote
	DashboardURLPrefix = "https://" + DashboardAddressPrefix

	// Docs page
	Docs = "http://wedeploy.com/docs/"

	// Hub for the system
	Hub = "http://api.wedeploy.io"

	// WeDeployImageTag is the WeDeploy image tag for docker
	WeDeployImageTag = "beta"

	// RequiresDockerConstraint semver version constraint
	RequiresDockerConstraint = ">= 1.12.0"

	// AnalyticsEndpoint for posting analytics events in bulk
	AnalyticsEndpoint = "https://cli-metrics.wedeploy.com/"

	// CloudRemote is the name for the default cloud for WeDeploy
	CloudRemote = "wedeploy"

	// LocalRemote is the local infrastructure remote name
	LocalRemote = "local"

	// LocalHTTPPort is the default port used to expose WeDeploy locally
	LocalHTTPPort = 80

	// LocalHTTPSPort is the default port used to expose WeDeploy locally over HTTPS
	LocalHTTPSPort = 443

	// SupportEmail value
	SupportEmail = "support@wedeploy.com"

	// StableReleaseChannel for the distribution of the CLI tool
	StableReleaseChannel = "stable"
)
