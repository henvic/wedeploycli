// Package extra resolves legal compliance for copy & pasted code.
// If you need something pretty small it is wise to copy & paste code.
// This is common good recommended practice on the Go community.
// "A little copying is better than a little dependency."
// Watch "Go Proverbs", by Rob Pike.
// https://go-proverbs.github.io
// https://www.youtube.com/watch?v=PAAkCSZUG1c&t=9m28s
package extra

// Licenses of extra dependency that are used on binary files,
// but instead of being availble on the vendor directory,
// are copied somewhere else.
var Licenses = []License{
	License{
		Name:        "moby",
		Package:     "github.com/moby/moby",
		Notes:       "modified",
		LicensePath: "extra/licenses/DOCKER_LICENSE", // Apache License 2.0
		// Files:
		// github.com/wedeploy/cli/namesgenerator
		// github.com/wedeploy/cli/templates
	},
}
