package errorhandling

var errorReasonMessage = messages{
	"unauthorized":              "Access is denied due to invalid credentials",
	"documentNotFound":          "Not found",
	"notFound":                  "Not found",
	"projectNotFound":           "Not found",
	"containerNotFound":         "Not found",
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
}

var errorReasonCommandMessageOverrides = map[string]messages{
	"run": messages{
		"typeNotFound":                  "Container type not found",
		"projectContainerQuotaExceeded": "Your quota for containers has exceeded",
		"exists":                        "Project is already linked",
	},
	"run stop": messages{
		"deleteProject": "Can not delete project",
	},
	"deploy": messages{
		"invalidDocumentValue": "Access denied to this project",
	},
	"undeploy": messages{
		"invalidContainer": "Not found",
	},
}
