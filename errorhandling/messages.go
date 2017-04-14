package errorhandling

var errorReasonMessage = messages{
	"unauthorized":                "Access is denied due to invalid credentials",
	"documentNotFound":            "Not found",
	"notFound":                    "Not found",
	"projectNotFound":             "Not found",
	"containerNotFound":           "Not found",
	"serviceNotFound":             "Not found",
	"badRequest":                  "The API request is invalid or improperly formed",
	"internalError":               "The request failed due to an internal error",
	"projectQuotaExceeded":        "Project quota exceeded",
	"invalidContainer":            "Invalid service",
	"invalidService":              "Invalid service",
	"invalidProject":              "Invalid project",
	"invalidAccountEmail":         "Invalid email account",
	"emailAlreadyExists":          "Email already exists",
	"invalidCollaboratorEmail":    "Invalid collaborator email",
	"invalidContainerId":          "Invalid service ID",
	"invalidServiceId":            "Invalid service ID",
	"containerAlreadyExists":      "Service already exists",
	"serviceAlreadyExists":        "Service already exists",
	"customDomainAlreadyExists":   "Custom domain already exists",
	"invalidProjectId":            "Invalid project ID",
	"projectAlreadyExists":        "Project already exists",
	"environmentVariableNotFound": "Environment variable not found",
}

var errorReasonCommandMessageOverrides = map[string]messages{
	"run": messages{
		"typeNotFound":                  "Container type not found",
		"projectContainerQuotaExceeded": "Your quota for services has exceeded",
		"projectServiceQuotaExceeded":   "Your quota for services has exceeded",
		"exists":                        "Project is already linked",
	},
	"deploy": messages{
		"invalidDocumentValue": "Access denied to this project",
	},
	"undeploy": messages{
		"deleteProject":    "Can not delete project",
		"invalidContainer": "Not found",
		"invalidService":   "Not found",
	},
	"login": messages{
		"validationError": "Invalid credentials",
	},
}
