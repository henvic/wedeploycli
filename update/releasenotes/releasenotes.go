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
		Description: "Added protection to avoid deploying content in sensitive directories such as the home directory. Minor improvements.",
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
		Description: `Added "we shell" command.`,
	},
	ReleaseNote{
		Version:     "1.4.5",
		Date:        "May 4th, 2018",
		Description: `Added release notes. Added timestamsp to "we log".`,
	},
	ReleaseNote{
		Version:     "1.4.6",
		Date:        "May 8th, 2018",
		Description: `Added support for using custom timezones (with environment variable TZ). Added support for deploying Git repositories. Fixing missing "error counter". Minor improvements.`,
	},
	ReleaseNote{
		Version:     "1.4.7",
		Date:        "May 13th, 2018",
		Description: `Improved user experience when running "we deploy". Minor improvements.`,
	},
	ReleaseNote{
		Version:     "1.4.8",
		Date:        "May 18th, 2018",
		Description: `Added support for piping username and password for "we login". Minor improvements.`,
	},
	ReleaseNote{
		Version:     "1.4.9",
		Date:        "May 30th, 2018",
		Description: `Fixing panic when using "we deploy --quiet". Fixing "we login" when using Git bash for Windows.`,
	},
	ReleaseNote{
		Version:     "1.4.10",
		Date:        "June 12th, 2018",
		Description: `Fix "we deploy" for Windows users whose usernames contains spaces. Minor improvements.`,
	},
	ReleaseNote{
		Version:     "1.5.0",
		Date:        "June 13th, 2018",
		Description: `Added the --skip-progress flag to "we deploy" and changed --quiet behavior to make it wait until deployment is finished.`,
	},
	ReleaseNote{
		Version:     "1.5.1",
		Date:        "June 15th, 2018",
		Description: `Improved output colors for requests when using --verbose. Minor improvements.`,
	},
	ReleaseNote{
		Version:     "1.5.2",
		Date:        "June 15th, 2018",
		Description: "Print friendly status text errors on API errors.",
	},
	ReleaseNote{
		Version:     "1.5.3",
		Date:        "June 19th, 2018",
		Description: "Minor improvements.",
	},
	ReleaseNote{
		Version:     "1.5.4",
		Date:        "June 19th, 2018",
		Description: "Improving error messages. Minor improvements.",
	},
	ReleaseNote{
		Version:     "1.5.5",
		Date:        "June 20th, 2018",
		Description: `Show current number of deployed instances on "we scale" and ask for service before prompting for number of instances on change.`,
	},
	ReleaseNote{
		Version:     "1.5.6",
		Date:        "June 22nd, 2018",
		Description: `Adding support to upcoming environment feature. Fixed flags on "we new". Minor improvements.`,
	},
	ReleaseNote{
		Version:     "1.5.7",
		Date:        "June 28th, 2018",
		Description: `Fix processing flags on command "we new project". Minor improvements.`,
	},
	ReleaseNote{
		Version:     "1.5.8",
		Date:        "July 4th, 2018",
		Description: "Fixed deployment upload failure feedback. Minor improvements.",
	},
	ReleaseNote{
		Version:     "1.5.9",
		Date:        "July 8th, 2018",
		Description: `Added "we list instances" command. Improved instance support. Minor improvements.`,
	},
	ReleaseNote{
		Version:     "1.5.10",
		Date:        "July 12th, 2018",
		Description: `Only print first 12 chars of instance ids. Autoconnect to instance on "we shell" when only one instance is running. Minor improvements.`,
	},
}
