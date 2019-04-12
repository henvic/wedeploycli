package releasenotes

import (
	"fmt"

	version "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/update/releasenotes"
)

var (
	from      string
	exclusive bool

	all bool
)

// Cmd for "liferay update release-notes"
var Cmd = &cobra.Command{
	Use:   "release-notes",
	Args:  cobra.NoArgs,
	RunE:  updateRun,
	Short: "Update Release Notes",
}

func updateRun(cmd *cobra.Command, args []string) error {
	notes := releasenotes.ReleaseNotes

	if all {
		printReleaseNotes(notes)
		return nil
	}

	switch from {
	case "", "master":
		notes = filterCurrentUpdate(notes)
	default:
		notes = filterUpdates(from, notes)
	}

	if len(notes) == 0 {
		return checkNoUpdateNotices(from)
	}

	printReleaseNotes(notes)

	return nil
}

func checkNoUpdateNotices(from string) error {
	fromV, fromErr := version.NewVersion(from)
	currentV, currentErr := version.NewVersion(defaults.Version)

	if from != "" && fromErr == nil && currentErr == nil && fromV.GreaterThan(currentV) {
		return fmt.Errorf("version on --from is %v, which is greater than the current version of this program: %v", fromV, currentV)
	}

	fmt.Println("No release notes found.")
	return nil
}

func filterCurrentUpdate(updates []releasenotes.ReleaseNote) []releasenotes.ReleaseNote {
	if defaults.Version == "master" {
		return []releasenotes.ReleaseNote{
			releasenotes.ReleaseNote{
				Version:     "master",
				Date:        "Future",
				Description: "You're using an unpublished version of this software.",
			},
		}
	}

	for _, u := range updates {
		if u.Version == defaults.Version {
			return []releasenotes.ReleaseNote{u}
		}
	}

	return nil
}

func filterUpdates(from string, notes []releasenotes.ReleaseNote) []releasenotes.ReleaseNote {
	var filtered []releasenotes.ReleaseNote
	fromV, fromErr := version.NewVersion(from)

	for c, n := range notes {
		noteV, noteErr := version.NewVersion(n.Version)

		if (fromErr == nil && noteErr == nil && fromV.LessThan(noteV)) || n.Version == from {
			filtered = notes[c:]
			break
		}
	}

	if exclusive && len(filtered) != 0 && filtered[0].Version == from {
		filtered = filtered[1:]
	}

	return filtered
}

func printReleaseNotes(notes []releasenotes.ReleaseNote) {
	for c, n := range notes {
		fmt.Printf("Version %v (%v)\n", n.Version, n.Date)
		fmt.Println(n.Description)

		if c != len(notes)-1 {
			fmt.Println()
		}
	}
}

func init() {
	Cmd.Flags().BoolVar(&all, "all", false, "See all notes")
	Cmd.Flags().StringVar(&from, "from", "", "See notes since a given version")
	Cmd.Flags().BoolVar(&exclusive, "exclusive", false, "Filter version equal to --from value")
}
