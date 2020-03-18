package deployremote

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wedeploy/cli/userhome"
)

var home = userhome.GetHomeDir()

var forbiddenLocations = []string{
	home,
	filepath.Join(home, "Downloads"),
	filepath.Join(home, "Desktop"),
	filepath.Join(home, "Documents"),
	filepath.Join(home, "Pictures"),
	filepath.Join(home, "Library"),
	filepath.Join(home, "Library", "Keychains"),
	filepath.Join(home, ".ssh"),
	filepath.Join(home, ".gnupg"),
	"/",
	"/etc",
	"/root",
	"/tmp",
	"/private",
	"/bin",
	"/home",
	"/mnt",
	"/Volumes",
	"/Users",
}

func getWorkingDirectory() (wd string, err error) {
	wd, err = os.Getwd()

	if err != nil {
		return "", err
	}

	for _, f := range forbiddenLocations {
		if strings.EqualFold(f, wd) {
			return "", fmt.Errorf("refusing to deploy from inside the top-level of \"%v\" because it might be unsafe", wd)
		}
	}

	return wd, nil
}
