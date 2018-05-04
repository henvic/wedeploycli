package releasenotes

// ReleaseNote for a given update.
type ReleaseNote struct {
	Version     string
	Date        string
	Description string
}

// ReleaseNotes for the updates.
var ReleaseNotes = []ReleaseNote{
	ReleaseNote{
		Version:     "1.1.5",
		Date:        "Oct 14th, 2017",
		Description: "Don't ask for authentication on `we console`. Minor improvements.",
	},
	ReleaseNote{
		Version:     "1.2.7",
		Date:        "Jan 2nd, 2018",
		Description: "Added prompt for projects and services on most commands. Added `we new` and `we open` commands. Minor improvements.",
	},
	ReleaseNote{
		Version:     "1.2.8",
		Date:        "Jan 5th, 2018",
		Description: "Adding protection to avoid deploying content in sensitive directories such as the home directory. Minor improvements.",
	},
	ReleaseNote{
		Version:     "1.2.9",
		Date:        "Jan 9th, 2018",
		Description: "Added prompt to verify project creation on `we deploy`. Improved removal protection, making you type the project or service name of the resource you want to remove on `we delete`. Minor improvements.",
	},

	ReleaseNote{
		Version:     "1.3.0",
		Date:        "Jan 12th, 2018",
		Description: "Improved user experience for the `we env` and `we domain` commands. Minor improvements.",
	},
	ReleaseNote{
		Version:     "1.3.1",
		Date:        "Jan 12th, 2018",
		Description: "Added support to applying environment variables from a file on `we env set`. Minor improvements.",
	},
	ReleaseNote{
		Version:     "1.3.2",
		Date:        "Jan 13th, 2018",
		Description: "Improved error messages on malformed wedeploy.json. Fixing bug on setting two environment variables at once. Added --replace flag to `we env set`. Minor improvements.",
	},
	ReleaseNote{
		Version:     "1.3.3",
		Date:        "Jan 14th, 2018",
		Description: "Validate wedeploy.json before trying to deploy. Added prompt for selecting or creating a project id on `we deploy`. Added commands `we list projects` and `we list services`. Added the --no-tty flag to make it easier to use the CLI programmatically. Minor improvements.",
	},
	ReleaseNote{
		Version:     "1.3.4",
		Date:        "Jan 15th, 2018",
		Description: "Fixed issue where deployment might never seem to terminate on CLI due to metadata type mismatch. Minor improvements.",
	},
	ReleaseNote{
		Version:     "1.3.5",
		Date:        "Feb 22nd, 2018",
		Description: "Fixing issue where nested services would be identified as services for the CLI. Fix skipping directories that have any files on the .gitignore list (instead of only the file itself). Minor improvements.",
	},
	ReleaseNote{
		Version:     "1.4.0",
		Date:        "Mar 6th, 2018",
		Description: `Making "we scale" work with no required arguments. Minor improvements.`,
	},
	ReleaseNote{
		Version:     "1.4.1",
		Date:        "Mar 6th, 2018",
		Description: `Minor improvements.`,
	},
	ReleaseNote{
		Version:     "1.4.2",
		Date:        "Mar 6th, 2018",
		Description: `Renaming "we env" with "we env-var". Stop allowing dashes on service ids. Minor improvements.`,
	},
	ReleaseNote{
		Version:     "1.4.4",
		Date:        "Mar 30th, 2018",
		Description: `Adding "we shell" command.`,
	},
}
