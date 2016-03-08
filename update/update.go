package update

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	gup "github.com/inconshreveable/go-update"
	"github.com/launchpad-project/api.go"
	"github.com/launchpad-project/cli/defaults"
)

// Release information
type Release struct {
	ID       string `json:"id"`
	Link     string `json:"link"`
	Version  string `json:"version"`
	Platform string `json:"platform"`
	Checksum string `json:"checksum"`
}

var (
	// ErrMasterVersion for when trying to run launchpad upgrade on a master distribution
	ErrMasterVersion = errors.New("You must upgrade a master version manually with git")

	// ErrPlatformUnsupported for when a release for the given platform was not found
	ErrPlatformUnsupported = errors.New("Build for your platform was not found")

	// ErrPermission for when there is no permission to replace the binary
	ErrPermission = errors.New("Can't replace Launchpad binary")

	signature = []byte(`-----BEGIN/END PUBLIC KEY-----`)
)

// GetLatestReleases lists the latest releases available for the given platform
func GetLatestReleases() []Release {
	var address = defaults.Endpoint + "/releases/dist/channel"
	var os = runtime.GOOS
	var arch = runtime.GOARCH

	var b = strings.NewReader(fmt.Sprintf(
		`{"filter": [{"platform": {"value": "%s/%s"}}]}`, os, arch,
	))

	req := launchpad.URL(address)
	req.Body(b)

	if err := req.Get(); err != nil {
		panic(err)
	}

	releases := *new([]Release)

	if err := req.DecodeJSON(&releases); err != nil {
		panic(err)
	}

	return releases
}

// ToLatest updates the Launchpad CLI to the latest version
func ToLatest() {
	if launchpad.Version == "master" {
		println(ErrMasterVersion.Error())
		os.Exit(1)
	}

	var releases = GetLatestReleases()

	if len(releases) == 0 {
		println("Releases not found.")
	}

	var next = releases[0]

	if next.Version == launchpad.Version {
		fmt.Println("Installed version " + launchpad.Version + " is already the latest version.")
		return
	}

	if err := Update(next); err != nil {
		panic(err)
	}
}

// Update updates the Launchpad CLI to the given release
func Update(release Release) error {
	resp, err := http.Get(release.Link)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	checksum, err := hex.DecodeString(release.Checksum)

	if err == nil {
		err = gup.Apply(resp.Body, gup.Options{
			Checksum:  checksum,
			Signature: signature,
		})
	}

	return err
}
