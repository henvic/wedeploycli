package errorhandling

var errorReasonMessage = messages{
	"documentNotFound": "Document not found.",
}

var errorReasonCommandMessageOverrides = map[string]messages{
	"list": messages{
		"projectNotFound": "Project not found.",
	},
}
