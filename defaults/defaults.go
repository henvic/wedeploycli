package defaults

var (
	// Version of the WeDeploy Project CLI tool
	Version = "master"

	// Build commit
	Build = ""

	// BuildTime is the time when the build was generated
	BuildTime = ""

	// DashboardAddress for the system (global)
	DashboardAddress = "dashboard.wedeploy.com"

	// Dashboard for the system (global)
	Dashboard = "http://" + DashboardAddress

	// OAuthTokenEndpoint for generating OAuth tokens
	OAuthTokenEndpoint = "http://auth.dashboard.wedeploy.com/oauth/token"

	// Docs page
	Docs = "http://wedeploy.com/docs/"

	// Hub for the system
	Hub = "http://api.dashboard.wedeploy.io"

	// WeDeployImageTag is the WeDeploy image tag for docker
	WeDeployImageTag = "beta"

	// RequiresDockerConstraint semver version constraint
	RequiresDockerConstraint = ">= 1.12.0"

	// AnalyticsEndpoint for posting analytics events in bulk
	AnalyticsEndpoint = "https://cli-metrics.wedeploy.com/"

	// DefaultCloudRemote is the name for the default cloud for WeDeploy
	DefaultCloudRemote = "wedeploy"

	// DefaultLocalRemote is the local infrastructure remote name
	DefaultLocalRemote = "local"
)
