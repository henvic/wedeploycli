package feedback

import "github.com/wedeploy/cli/activities"

// ValidBuildTransitions for build states (from -> to)
var ValidBuildTransitions = map[string][]string{
	activities.BuildFailed: []string{
		activities.BuildSucceeded,
		activities.DeployFailed,
		activities.DeployCanceled,
		activities.DeployTimeout,
		activities.DeployRollback,
		activities.DeploySucceeded,
	},
	activities.BuildSucceeded: []string{
		activities.DeployFailed,
		activities.DeployCanceled,
		activities.DeployTimeout,
		activities.DeployRollback,
		activities.DeploySucceeded,
	},
}

// ValidDeployTransitions for deploy states (from -> to)
var ValidDeployTransitions = map[string][]string{
	activities.BuildFailed: []string{
		activities.BuildSucceeded,
		activities.DeployFailed,
		activities.DeployCanceled,
		activities.DeployTimeout,
		activities.DeployRollback,
		activities.DeploySucceeded,
	},
	activities.BuildSucceeded: []string{
		activities.DeployFailed,
		activities.DeployCanceled,
		activities.DeployTimeout,
		activities.DeployRollback,
		activities.DeploySucceeded,
	},
	activities.DeployFailed: []string{
		activities.DeploySucceeded,
	},
	activities.DeployCanceled: []string{
		activities.DeploySucceeded,
	},
	activities.DeployTimeout: []string{
		activities.DeploySucceeded,
	},
	activities.DeployRollback: []string{
		activities.DeploySucceeded,
	},
	activities.DeploySucceeded: []string{},
}
