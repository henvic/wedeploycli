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
		Name:        "Go",
		Package:     "github.com/golang/go",
		LicensePath: "extra/licenses/GOLANG_LICENSE",
	},
	License{
		Name:        "moby",
		Package:     "github.com/moby/moby",
		Notes:       "modified",
		LicensePath: "extra/licenses/DOCKER_LICENSE", // Apache License 2.0
		// Files:
		// github.com/henvic/wedeploycli/namesgenerator
		// github.com/henvic/wedeploycli/templates
	},
	License{
		Name:        "color",
		Package:     "github.com/fatih/color",
		Notes:       "heavily modified",
		LicensePath: "extra/licenses/COLOR_LICENSE", // MIT
		// Files:
		// github.com/henvic/wedeploycli/color
	},
	License{
		Name:        "figures",
		Package:     "github.com/sindresorhus/figures",
		Notes:       "heavily simplified",
		LicensePath: "extra/licenses/FIGURES_LICENSE", // MIT
		// Files:
		// github.com/henvic/wedeploycli/figures
	},
	License{
		Name:        "buffruneio",
		Package:     "github.com/pelletier/go-buffruneio",
		LicensePath: "extra/licenses/BUFFRUNEIO_LICENSE", // MIT
		// Files:
		// github.com/pelletier/go-buffruneio (vendored)
		// See https://github.com/pelletier/go-buffruneio/issues/6
	},
}
