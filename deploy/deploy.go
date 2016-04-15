package deploy

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/launchpad-project/api.go"
	"github.com/launchpad-project/cli/apihelper"
	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/containers"
	"github.com/launchpad-project/cli/hooks"
	"github.com/launchpad-project/cli/pod"
	"github.com/launchpad-project/cli/progress"
	"github.com/launchpad-project/cli/projects"
	"github.com/launchpad-project/cli/verbose"
)

// Deploy holds the information of a POD to be zipped or deployed
type Deploy struct {
	Container     containers.Container
	ContainerPath string
	PackageSize   int64
	progress      *progress.Bar
}

// DeployFlags modifiers
type DeployFlags struct {
	Hooks bool
}

// ErrDeploy is a generic error triggered when any deploy error happens
var ErrDeploy = errors.New("Error during deploy")

// All deploys a list of containers on the given context
func All(list []string, df *DeployFlags) (err error) {
	var wg sync.WaitGroup
	var el []error

	wg.Add(len(list))

	for _, container := range list {
		go func(c string) {
			el = append(el, Only(c, df))
			wg.Done()
		}(container)
	}

	wg.Wait()
	progress.Stop()

	for _, e := range el {
		if e == nil {
			continue
		}

		println(e.Error())
		err = ErrDeploy
	}

	return err
}

// Only PODify a container and deploys it to Launchpad
func Only(container string, df *DeployFlags) error {
	var deploy, err = New(container)

	if err != nil {
		return err
	}

	var projectID = config.Stores["project"].Get("id")

	err = projects.Validate(projectID)

	switch err {
	case projects.ErrProjectAlreadyExists:
		break
	case nil:
		err = projects.Create(projectID, config.Stores["project"].Get("name"))

		if err != nil {
			return err
		} else {
			fmt.Println("New project created")
		}
	default:
		return err
	}

	err = containers.Validate(projectID, deploy.Container.ID)

	switch err {
	case containers.ErrContainerAlreadyExists:
		err = nil
		break
	case nil:
		err = containers.Install(projectID, deploy.Container)

		if err != nil {
			return err
		} else {
			fmt.Println("New container installed")
		}
	default:
		return err
	}

	var containerHooks = deploy.Container.Hooks

	if df.Hooks && containerHooks != nil && containerHooks.BeforeDeploy != "" {
		err = hooks.Run(containerHooks.BeforeDeploy)
	}

	if err == nil {
		err = deploy.Only()
	}

	if err == nil && df.Hooks && containerHooks != nil && containerHooks.AfterDeploy != "" {
		err = hooks.Run(containerHooks.AfterDeploy)
	}

	return err
}

// New Deploy instance
func New(container string) (*Deploy, error) {
	var deploy = &Deploy{
		ContainerPath: path.Join(config.Context.ProjectRoot, container),
		progress:      progress.New(container),
	}

	var err = containers.GetConfig(deploy.ContainerPath, &deploy.Container)

	return deploy, err
}

// Zip packages a POD to a .pod package
func Zip(dest, container string) error {
	var deploy, err = New(container)

	if err == nil {
		err = deploy.Zip(dest)
	}

	return err
}

// Deploy POD to Launchpad
func (d *Deploy) Deploy(pod string) (err error) {
	var projectID = config.Stores["project"].Get("id")
	var u = path.Join("api/push", projectID, d.Container.ID)
	var req = apihelper.URL(u)
	var file io.Reader

	apihelper.Auth(req)

	w := &writeCounter{
		progress: d.progress,
		Size:     uint64(d.PackageSize),
	}

	d.progress.Reset("Uploading", "")
	file, err = os.Open(pod)

	if err == nil {
		req.Body(io.TeeReader(file, w))
	}

	if err == nil {
		err = apihelper.Validate(req, req.Post())
	}

	if err == nil || err == launchpad.ErrUnexpectedResponse {
		d.progress.Append = fmt.Sprintf(
			"%s (Complete)",
			humanize.Bytes(uint64(d.PackageSize)))
		d.progress.Set(progress.Total)
	}

	if err == nil {
		fmt.Printf(fmt.Sprintf("Ready! %v.%v.liferay.io\n", d.Container.ID, projectID))
	}

	return err
}

// Only PODify a container and deploys it to Launchpad
func (d *Deploy) Only() error {
	tmp, err := ioutil.TempFile(os.TempDir(), "launchpad-cli")

	err = d.Zip(tmp.Name())

	if err == nil {
		err = d.Deploy(tmp.Name())
	}

	if tmp != nil {
		os.Remove(tmp.Name())
	}

	return err
}

// Zip packages a POD to a .pod package
func (d *Deploy) Zip(dest string) (err error) {
	var c containers.Container
	d.progress.Reset("Zipping", "")
	dest, _ = filepath.Abs(dest)

	c.DeployIgnore = append(c.DeployIgnore, pod.CommonIgnorePatterns...)

	// avoid zipping itself 'til starvation
	c.DeployIgnore = append(c.DeployIgnore, dest)

	d.PackageSize, err = pod.Compress(
		dest,
		d.ContainerPath,
		c.DeployIgnore,
		d.progress)

	if err == nil {
		d.progress.Set(progress.Total)
	}

	verbose.Debug("Saving container to", dest)

	return err
}




}
