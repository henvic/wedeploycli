package defaults

var (
	// Version of the WeDeploy Project CLI tool
	Version = "master"

	// Build commit
	Build = ""

	// BuildTime is the time when the build was generated
	BuildTime = ""

	// Hub for the system
	Hub = "http://api.dashboard.wedeploy.io"

	// WeDeployImageTag is the WeDeploy image tag for docker
	WeDeployImageTag = "latest"

	// RequiresDockerConstraint semver version constraint
	RequiresDockerConstraint = ">= 1.12.0"

	// AnalyticsEndpoint for posting analytics events in bulk
	AnalyticsEndpoint = "https://cli-metrics.wedeploy.com/"

	// DefaultCloudRemote is the name for the default cloud for WeDeploy
	DefaultCloudRemote = "wedeploy"

	// DefaultLocalRemote is the local infrastructure remote name
	DefaultLocalRemote = "local"
)
