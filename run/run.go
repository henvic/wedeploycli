package run

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/errwrap"
	semver "github.com/hashicorp/go-version"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/exechelper"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/prompt"
	"github.com/wedeploy/cli/verbose"
	"golang.org/x/crypto/ssh/terminal"
)

// WeDeployImage is the docker image for the WeDeploy infrastructure
var WeDeployImage = "wedeploy/local:" + defaults.WeDeployImageTag

var bin = "docker"

type tcpPortsStruct []int

func (t tcpPortsStruct) getAvailability() (all bool, notAvailable []int) {
	all = true
	for _, k := range t {
		// there is a small chance of a port being in use by a process, but not
		// responding. We ignore this risk here for simplicity.
		var con, err = net.Dial("tcp", fmt.Sprintf(":%v", k))

		if con != nil {
			_ = con.Close()
			all = false
			notAvailable = append(notAvailable, k)
			continue
		}

		switch err.(type) {
		case *net.OpError:
			// ignore error as we want the port to be free
			// this is not 100% bullet-proof, but good enough for our needs
			continue
		default:
			verbose.Debug("Ignoring unexpected error", err)
		}
	}

	return all, notAvailable
}

func (t tcpPortsStruct) expose() []string {
	var ports []string
	for _, k := range t {
		ports = append(ports, "-p", fmt.Sprintf("%v:%v", k, k))
	}

	return ports
}

// GetWeDeployHost gets the WeDeploy infrastructure host
// This is a temporary solution and it is NOT reliable
func GetWeDeployHost() (string, error) {
	var params = []string{
		"network", "inspect", "bridge", "--format", `{{(index (index .IPAM.Config 0) "Gateway")}}`,
	}

	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(params, " ")))
	var gatewayBuf bytes.Buffer
	var gateway = exec.Command(bin, params...)
	gateway.Stderr = os.Stderr
	gateway.Stdout = &gatewayBuf

	if err := gateway.Run(); err != nil {
		return "", errwrap.Wrapf("Can not get docker network bridge gateway: {{err}}", err)
	}

	var address = gatewayBuf.String()
	verbose.Debug("docker network bridge gateway address:", address)
	return strings.TrimSpace(address), nil
}

// StopOutdatedImage stops the WeDeploy infrastructure if outdated
func StopOutdatedImage(nextImage string) error {
	// don't try to stop if docker isn't installed yet
	if !existsDependency(bin) {
		return nil
	}

	var dm = &DockerMachine{}

	if err := dm.LoadDockerInfo(); err != nil {
		return err
	}

	if dm.Container == "" {
		return nil
	}

	if nextImage == WeDeployImage && nextImage != defaults.WeDeployImageTag {
		verbose.Debug("Continuing update without stopping: same infrastructure version detected.")
		return nil
	}

	println("New WeDeploy infrastructure image available.")
	println("The infrastructure must be stopped before updating the CLI tool.")

	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		return errors.New("No terminal (/dev/tty) detected for asking to stop the infrastructure. Exiting.")
	}

	if nextImage == "latest" {
		println("Notice: Updating to latest always requires WeDeploy infrastructure to be turned off.")
	}

	var q, err = prompt.Prompt("Stop WeDeploy to allow update [yes]")

	if err != nil {
		return err
	}

	if q != "" && q != "y" && q != "yes" {
		return errors.New("Can not update image while running an old version of the infrastructure.")
	}

	return cleanupEnvironment()
}

func checkDockerAvailable() error {
	if !existsDependency(bin) {
		return errors.New("Docker is not installed. Download it from http://docker.com/")
	}

	var params = []string{
		"version", "--format", "{{.Client.Version}}",
	}

	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(params, " ")))
	var versionErrBuf bytes.Buffer
	var versionBuf bytes.Buffer
	var version = exec.Command(bin, params...)
	version.Stderr = &versionErrBuf
	version.Stdout = &versionBuf

	err := version.Run()

	switch {
	case err != nil && strings.Contains(versionErrBuf.String(),
		"Is the docker daemon running on this host?\n"):
		return errors.New(strings.TrimSpace(versionErrBuf.String()))
	case err != nil:
		print(versionErrBuf.String())
		return errwrap.Wrapf("Cannot check docker version: {{err}}", err)
	}

	v := strings.TrimSpace(versionBuf.String())
	installedDockerVersion, err := semver.NewVersion(v)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing docker [semantic] version: %v (ignoring)\n", err)
		return nil
	}

	constraints, err := semver.NewConstraint(defaults.RequiresDockerConstraint)

	if err != nil {
		panic(err)
	}

	if !constraints.Check(installedDockerVersion) {
		return fmt.Errorf(`docker version too old: ` +
			color.Format(color.FgHiRed, `%v`, installedDockerVersion) +
			", required is " +
			color.Format(color.FgHiRed, `%v`,
				defaults.RequiresDockerConstraint) + `
Update it or download a new version from http://docker.com/
	If this doesn't work:
	1) check for multiple older docker versions on your system
	2) if you find them, backup any containers or settings you need
	3) stop and uninstall all docker instances ` +
			color.Format(color.Bold, "until the docker command fails") + "\n" +
			`	4) install docker again`)
	}

	return nil
}

func getWeDeployHost() string {
	var address, err = GetWeDeployHost()

	if err != nil {
		println("Could not find a suitable host.")
		println("To use \"we run\" you need a suitable docker network interface on.")
		println(err.Error())
		os.Exit(1)
	}

	return address
}

func startCmd(args ...string) (dockerContainer string, err error) {
	var docker = exec.Command(bin, args...)
	var dockerContainerBuf bytes.Buffer
	docker.Stderr = os.Stderr
	docker.Stdout = &dockerContainerBuf

	if err = docker.Run(); err != nil {
		return "", errwrap.Wrapf("docker run error: {{err}}", err)
	}

	return strings.TrimSpace(dockerContainerBuf.String()), err
}

func unlinkProjects() error {
	verbose.Debug("Unlinking projects")

	list, err := projects.List(context.Background())

	if err != nil {
		return errwrap.Wrapf("Can not list projects for unlinking: {{err}}", err)
	}

	for _, p := range list {
		if err := projects.Unlink(context.Background(), p.ID); err != nil {
			fmt.Fprintf(os.Stderr, "Unlinking project %v error: %v\n", p.ID, err)
		}
	}

	return err
}

func cleanupEnvironment() error {
	verbose.Debug("Cleaning up processes and containers.")

	if err := stopContainers(); err != nil {
		return err
	}

	if err := rmContainers(); err != nil {
		return err
	}

	if err := rmOldInfrastructureImages(); err != nil {
		return err
	}

	verbose.Debug("End of environment clean up.")

	ids, err := getDockerContainers(true)

	if err != nil {
		return errwrap.Wrapf("Can not verify containers are down: {{err}}", err)
	}

	if len(ids) != 0 {
		err = fmt.Errorf("Containers still up after shutdown procedure: %v", ids)
	}

	return err
}

func stopContainers() error {
	verbose.Debug("Trying to stop WeDeploy containers and infrastructure containers.")
	var ids, err = getDockerContainers(true)

	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	var params = []string{"stop"}
	params = append(params, ids...)
	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(params, " ")))
	var stop = exec.Command(bin, params...)
	exechelper.AddCommandToNewProcessGroup(stop)
	stop.Stderr = os.Stderr

	switch err = stop.Run(); err.(type) {
	case nil:
		return nil
	case *exec.ExitError:
		return errwrap.Wrapf("warning: still stopping WeDeploy on background: {{err}}", err)
	default:
		return errwrap.Wrapf("docker stop error: {{err}}", err)
	}
}

func rmContainers() error {
	var ids, err = getDockerContainers(false)

	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	var params = []string{"rm", "--force"}
	params = append(params, ids...)
	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(params, " ")))
	var rm = exec.Command(bin, params...)
	exechelper.AddCommandToNewProcessGroup(rm)
	rm.Stderr = os.Stderr

	if err = rm.Run(); err != nil {
		return errwrap.Wrapf("Error trying to remove containers: {{err}}", err)
	}

	return err
}

func rmOldInfrastructureImages() error {
	var ids, err = getOldInfrastructureImages()

	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	var params = []string{"rmi"}
	params = append(params, ids...)
	params = append(params, "--force")
	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(params, " ")))
	var rm = exec.Command(bin, params...)
	exechelper.AddCommandToNewProcessGroup(rm)
	rm.Stderr = os.Stderr

	if err = rm.Run(); err != nil {
		return errwrap.Wrapf("Error trying to remove images: {{err}}", err)
	}

	return err
}

func getDockerContainers(onlyRunning bool) (cids []string, err error) {
	cids, err = getContainersByLabel("com.wedeploy.container.type", onlyRunning)

	if err != nil {
		return []string{}, err
	}

	idsInfra, err := getContainersByLabel("com.wedeploy.project.infra", onlyRunning)

	if err != nil {
		return []string{}, err
	}

	return append(cids, idsInfra...), err
}

func getContainersByLabel(label string, onlyRunning bool) (cs []string, err error) {
	var params = []string{
		"ps", "--filter", "label=" + label, "--quiet", "--no-trunc",
	}

	if !onlyRunning {
		params = append(params, "--all")
	}

	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(params, " ")))
	var list = exec.Command(bin, params...)
	exechelper.AddCommandToNewProcessGroup(list)
	var buf bytes.Buffer
	list.Stderr = os.Stderr
	list.Stdout = &buf

	if err := list.Run(); err != nil {
		return cs, errwrap.Wrapf("Can not get containers list: {{err}}", err)
	}

	return strings.Fields(buf.String()), nil
}

func getOldInfrastructureImages() ([]string, error) {
	var params = []string{
		"images",
		"wedeploy/local",
		"--format",
		"{{.Tag}}\t{{.ID}}",
		"--no-trunc",
	}

	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(params, " ")))
	var list = exec.Command(bin, params...)
	exechelper.AddCommandToNewProcessGroup(list)
	var buf bytes.Buffer
	list.Stderr = os.Stderr
	list.Stdout = &buf

	if err := list.Run(); err != nil {
		return []string{}, err
	}

	var images = strings.Split(buf.String(), "\n")
	var imageHashes = []string{}

	for _, ia := range images {
		var i = strings.Fields(ia)
		if len(i) == 2 && (i[0] != "latest" && i[0] != defaults.WeDeployImageTag) {
			imageHashes = append(
				imageHashes,
				strings.TrimSuffix(i[1], "sha256:"))
		}
	}

	return imageHashes, nil
}

func existsDependency(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
