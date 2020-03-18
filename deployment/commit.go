package deployment

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/henvic/wedeploycli/defaults"
	"github.com/henvic/wedeploycli/deployment/internal/repodiscovery"
	"github.com/henvic/wedeploycli/deployment/internal/repodiscovery/tiny"
	"github.com/henvic/wedeploycli/services"
	"github.com/henvic/wedeploycli/verbose"
	git "gopkg.in/src-d/go-git.v4"
)

func (d *Deploy) printPackageSize() {
	var s uint64

	// not the best thing to do in terms of performance, by the way
	// https://github.com/golang/go/issues/16399
	f := func(path string, info os.FileInfo, err error) error {
		if info != nil {
			s += uint64(info.Size())
		}

		return err
	}

	pkg := filepath.Join(d.workDir, git.GitDirName)

	if err := filepath.Walk(pkg, f); err != nil {
		verbose.Debug("can't get deployment size correctly:", err)
	}

	d.watch.PrintPackageSize(s)
}

// Info about the deployment.
type Info struct {
	CLIVersion string `json:"cliVersion,omitempty"`
	Time       string `json:"time,omitempty"`
	Deploy     bool   `json:"deploy,omitempty"`

	Repositories []repodiscovery.Repository `json:"repos,omitempty"`
	Repoless     []string                   `json:"repoless,omitempty"`

	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// Info about the deployment.
func (d *Deploy) Info() string {
	version := fmt.Sprintf("%s %s/%s",
		defaults.Version,
		runtime.GOOS,
		runtime.GOARCH)

	repositories, repoless := getProjectOrServiceInfo(d.Path, d.Services)

	di := Info{
		CLIVersion:   version,
		Time:         time.Now().Format(time.RubyDate),
		Deploy:       !d.OnlyBuild,
		Repositories: repositories,
		Repoless:     repoless,
		Metadata:     d.Metadata,
	}

	bdi, err := json.Marshal(tiny.Convert(tiny.Info(di)))

	if err != nil {
		verbose.Debug(err)
	}

	return string(bdi)
}

func getProjectOrServiceInfo(path string, s services.ServiceInfoList) ([]repodiscovery.Repository, []string) {
	repositories, repoless := getInfo(path, nil)

	if len(repositories) != 0 {
		return repositories, repoless
	}

	var err error
	path, err = filepath.Abs(filepath.Join(path, ".."))

	if err != nil {
		verbose.Debug("cannot go back one directory:", err)
		return nil, nil
	}

	return getInfo(filepath.Join(path, ".."), nil)
}

func getInfo(path string, s services.ServiceInfoList) ([]repodiscovery.Repository, []string) {
	rd := repodiscovery.Discover{
		Path:     path,
		Services: s,
	}

	repositories, repoless, err := rd.Run()

	if err != nil {
		verbose.Debug(err)
		return nil, nil
	}

	if len(repositories) == 0 {
		rd = repodiscovery.Discover{
			Path:     filepath.Join(path, ".."),
			Services: s,
		}

		repositories, _, err = rd.Run()

		if err != nil {
			verbose.Debug(err)
			return nil, nil
		}
	}

	return repositories, repoless
}
