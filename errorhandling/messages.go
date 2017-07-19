package errorhandling

var errorReasonMessage = messages{
	"restricted":                     "Access is restricted to collaborators",
	"unauthorized":                   "Access is denied due to invalid credentials",
	"documentNotFound":               "Not found",
	"notFound":                       "Not found",
	"projectNotFound":                "Not found",
	"serviceNotFound":                "Not found",
	"badRequest":                     "The API request is invalid or improperly formed",
	"internalError":                  "The request failed due to an internal error",
	"projectQuotaExceeded":           "Project quota exceeded",
	"invalidParameter":               "Invalid parameter",
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
}

var errorReasonCommandMessageOverrides = map[string]messages{
	"deploy": messages{
		"invalidDocumentValue": "Access denied to this project",
	},
	"remove": messages{
		"deleteProject":  "Can not delete project",
		"invalidService": "Not found",
	},
	"login": messages{
		"validationError": "Invalid credentials",
	},
}
