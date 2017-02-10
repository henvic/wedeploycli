package errorhandling

var errorReasonMessage = messages{
	"unauthorized":              "Access is denied due to invalid credentials",
	"documentNotFound":          "Not found",
	"notFound":                  "Not found",
	"badRequest":                "The API request is invalid or improperly formed",
	"internalError":             "The request failed due to an internal error",
	"projectQuotaExceeded":      "Project quote exceeded",
	"invalidContainer":          "Invalid container",
	"invalidProject":            "Invalid project",
	"invalidAccountEmail":       "Invalid email account",
	"emailAlreadyExists":        "Email already exists",
	"invalidCollaboratorEmail":  "Invalid collaborator email",
	"invalidContainerId":        "Invalid container ID",
	"containerAlreadyExists":    "Container already exists",
	"customDomainAlreadyExists": "Custom domain already exists",
	"invalidProjectId":          "Invalid project ID",
	"projectAlreadyExists":      "Project already exists",
	"projectNotFound":           "Project not found",
	"containerNotFound":         "Container not found",
}

var errorReasonCommandMessageOverrides = map[string]messages{
	"dev": messages{
		"typeNotFound":                  "Container type not found",
		"projectContainerQuotaExceeded": "Your quota for containers has exceeded",
		"exists":                        "Project is already linked",
	},
	"dev stop": messages{
		"deleteProject": "Can not delete project",
	},
	"domain": messages{
		"notFound":         "Project not found",
		"documentNotFound": "Project not found",
	},
	"env": messages{
		"notFound":         "Container not found",
		"documentNotFound": "Container not found",
	},
	"list": messages{
		"documentNotFound": "Not found",
		"notFound":         "Not found",
	},
	"undeploy": messages{
		"notFound":         "Project not found",
		"invalidContainer": "Container not found",
	},
}
