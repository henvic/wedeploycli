package errorhandling

var errorReasonMessage = messages{
	"unauthorized":     "Access is denied due to invalid credentials",
	"documentNotFound": "Document not found",
	"notFound": "The requested operation failed because a resource associated" +
		" with the request could not be found",
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
	"link": messages{
		"typeNotFound":                  "Container type not found",
		"projectContainerQuotaExceeded": "Your quota for containers has exceeded",
	},
	"unlink": messages{
		"documentNotFound": "Can not find project or container",
		"deleteProject":    "Can not delete project",
	},
	"list": messages{
		"projectNotFound":  "Project not found",
		"documentNotFound": "Project not found",
	},
	"log": messages{
		"notFound": "Log not found",
	},
}
