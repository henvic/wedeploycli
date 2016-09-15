package run

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/prompt"
	"github.com/wedeploy/cli/verbose"
)

// ErrHostNotFound is used when host is not found
var ErrHostNotFound = errors.New("You need to be connected to a network.")

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
	var addrs, err = net.InterfaceAddrs()

	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		var ip = addr.(*net.IPNet)

		if !ip.IP.IsLoopback() && ip.IP.To4() != nil {
			return strings.SplitN(ip.String(), "/", 2)[0], nil
		}
	}

	return "", ErrHostNotFound
}

// StopOutdatedImage stops the WeDeploy infrastructure if outdated
func StopOutdatedImage(nextImage string) error {
	// don't try to stop if docker isn't installed yet
	if !existsDependency(bin) {
		return nil
	}

	var dm = &DockerMachine{}

	dm.LoadDockerInfo()

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

	var q = prompt.Prompt("Stop WeDeploy to allow update [yes]")

	if q != "" && q != "y" && q != "yes" {
		return errors.New("Can't update image while running an old version of the infrastructure.")
	}

	return cleanupEnvironment()
}

func checkDockerExists() error {
	if !existsDependency(bin) {
		return errors.New("Docker is not installed. Download it from http://docker.com/")
	}

	return nil
}

func getWeDeployHost() string {
	var address, err = GetWeDeployHost()

	if err != nil {
		println("Could not find a suitable host.")
		println("To use we run you need a suitable network interface on.")
		println(err.Error())
		os.Exit(1)
	}

	return address
}

func runWait(container string) (*os.Process, error) {
	var c = exec.Command(bin, "wait", container)
	tryAddCommandToNewProcessGroup(c)
	var err = c.Start()

	return c.Process, err
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

	list, err := projects.List()

	if err != nil {
		return errwrap.Wrapf("Can't list projects for unlinking: {{err}}", err)
	}

	for _, p := range list {
		if err := projects.Unlink(p.ID); err != nil {
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
		return errwrap.Wrapf("Can't verify containers are down: {{err}}", err)
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
	tryAddCommandToNewProcessGroup(stop)
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
	tryAddCommandToNewProcessGroup(rm)
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
	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(params, " ")))
	var rm = exec.Command(bin, params...)
	tryAddCommandToNewProcessGroup(rm)
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
	tryAddCommandToNewProcessGroup(list)
	var buf bytes.Buffer
	list.Stderr = os.Stderr
	list.Stdout = &buf

	if err := list.Run(); err != nil {
		return cs, errwrap.Wrapf("Can't get containers list: {{err}}", err)
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
	tryAddCommandToNewProcessGroup(list)
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
