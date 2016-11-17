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
)
