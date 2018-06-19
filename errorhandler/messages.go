package errorhandler

import "github.com/wedeploy/cli/defaults"

var errorReasonMessage = messages{
	"restricted":                     "Access is restricted to collaborators",
	"unauthorized":                   "Access is denied due to invalid credentials",
	"invalidCredentials":             "Access is denied due to invalid credentials",
	"documentNotFound":               "Not found",
	"notFound":                       "Not found",
	"badRequest":                     "The API request is invalid or improperly formed",
	"internalError":                  "The request failed due to an internal error",
	"projectQuotaExceeded":           "Project quota exceeded",
	"exceededProjectMaximum":         "Project quota exceeded",
	"invalidService":                 "Invalid service",
	"invalidProject":                 "Invalid project",
	"invalidAccountEmail":            "Invalid email account",
	"emailAlreadyExists":             "Email already exists",
	"emailInvalidOrAlreadyBeingUsed": "Email is invalid or already being used",
	"invalidCollaboratorEmail":       "Invalid collaborator email",
	"invalidServiceId":               "Invalid service ID",
	"serviceAlreadyExists":           "Service already exists",
	"customDomainAlreadyExists":      "Custom domain already exists",
	"invalidProjectId":               "Invalid project ID",
	"projectAlreadyExists":           "Project already exists",
	"environmentVariableNotFound":    "Environment variable not found",
	"exceededPlanMaximum":            "You've reached your plan limits. Upgrade at " + defaults.PlanUpgradeURL,

	// special case: invalidParameter always uses the message from the context unless it is not set
	"invalidParameter": `{{if .message}}{{.message}}{{else}}Invalid value "{{.value}}" for parameter "{{.param}}"{{end}}`,
}

var errorReasonCommandMessageOverrides = map[string]messages{
	"deploy": messages{
		"invalidDocumentValue": "Access denied to this project",
		"restricted": `Looks like this project already exists and you don't have access to it.
Please try another project ID or make sure someone adds you as a collaborator`,
	},
	"remove": messages{
		"deleteProject":  "Can not delete project",
		"invalidService": "Not found",
	},
	"login": messages{
		"validationError": "Invalid credentials",
	},
}
