package run

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/prompt"
	"github.com/wedeploy/cli/verbose"
)

const (
	// WarmupOn symbol
	WarmupOn = '○'

	// WarmupOff symbol
	WarmupOff = '●'
)

// ErrHostNotFound is used when host is not found
var ErrHostNotFound = errors.New("You need to be connected to a network.")

// WeDeployImage is the docker image for the WeDeploy infrastructure
var WeDeployImage = "wedeploy/local:" + defaults.WeDeployImageTag

var bin = "docker"

var dockerLatestImageTag = "latest"

// Flags modifiers
type Flags struct {
	Debug    bool
	Detach   bool
	DryRun   bool
	ViewMode bool
	NoUpdate bool
}

// DockerMachine for the run command
type DockerMachine struct {
	Container string
	Image     string
	Flags     Flags
	upTime    time.Time
	// waitProcess is the "docker wait" PID
	waitProcess    *os.Process
	livew          *uilive.Writer
	tickerd        chan bool
	end            chan bool
	started        chan bool
	selfStopSignal bool
	tcpPorts       tcpPortsStruct
}

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

// Run runs the WeDeploy infrastructure
func Run(flags Flags) error {
	if err := checkDockerExists(); err != nil {
		return err
	}

	var dm = &DockerMachine{
		Flags: flags,
	}

	return dm.Run()
}

// Stop stops the WeDeploy infrastructure
func Stop() error {
	if err := checkDockerExists(); err != nil {
		return err
	}

	var dm = &DockerMachine{}
	return dm.Stop()
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

	if nextImage == WeDeployImage && nextImage != dockerLatestImageTag {
		verbose.Debug("Continuing update without stopping: same infrastructure version detected.")
		return nil
	}

	println("New WeDeploy infrastructure image available.")
	println("The infrastructure must be stopped before updating the CLI tool.")

	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		return errors.New("No terminal (/dev/tty) detected for asking to stop the infrastructure. Exiting.")
	}

	if nextImage == dockerLatestImageTag {
		println("Notice: Updating to latest always requires WeDeploy infrastructure to be turned off.")
	}

	var q = prompt.Prompt("Stop WeDeploy to allow update [yes]")

	if q != "" && q != "y" && q != "yes" {
		return errors.New("Can't update image while running an old version of the infrastructure.")
	}

	return cleanupEnvironment()
}

// Run executes the WeDeploy infraestruture
func (dm *DockerMachine) Run() (err error) {
	dm.prepare()

	if dm.Flags.Detach {
		dm.end <- true
	}

	if len(dm.Container) == 0 && dm.Flags.ViewMode {
		return errors.New("View mode is not available: WeDeploy is shutdown.")
	}

	var already = len(dm.Container) != 0 && !dm.Flags.DryRun

	if already {
		fmt.Println("WeDeploy is already running.")

		if dm.Flags.Debug {
			fmt.Fprintf(os.Stderr, "Can't expose debug ports because system is already up.\n")
		}
	} else if err = cleanupEnvironment(); err != nil {
		return err
	}

	dm.maybeStopListener()

	if !already {
		if err = dm.start(); err != nil {
			return err
		}
	}

	dm.maybeWaitEnd()
	dm.started <- true
	go dm.waitReadyState()
	<-dm.end
	return nil
}

// Stop stops the machine
func (dm *DockerMachine) Stop() error {
	dm.LoadDockerInfo()

	if dm.Container == "" {
		verbose.Debug("No infrastructure container detected.")
	}

	if err := cleanupEnvironment(); err != nil {
		return err
	}

	// try to terminate wait process child by sending a SIGTERM
	if dm.waitProcess != nil {
		if err := dm.waitProcess.Signal(syscall.SIGTERM); err != nil {
			return errwrap.Wrapf("Can't terminate docker wait process: {{err}}", err)
		}
	}

	return nil
}

func (dm *DockerMachine) setupPorts() {
	dm.tcpPorts = tcpPortsStruct{
		80,
		8080,
		24224,
	}

	if dm.Flags.Debug {
		dm.tcpPorts = append(dm.tcpPorts,
			5001,
			5005,
			8001,
			8500,
			9200)
	}
}

func (dm *DockerMachine) checkPortsAreAvailable() error {
	var all, notAvailable = dm.tcpPorts.getAvailability()

	if all {
		return nil
	}

	var s = "Can't start. The following network ports must be available:"

	for _, port := range notAvailable {
		s += fmt.Sprintf("%v\n", port)
	}

	s += "Sometimes docker doesn't free up ports properly.\n" +
		"If you don't know why these ports are not available, try restarting docker.\n"

	return errors.New(s)
}

func (dm *DockerMachine) waitEnd() {
	var p, err = runWait(dm.Container)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Running wait error: %v", err)
	}

	verbose.Debug("docker wait process pid:", p.Pid)
	dm.waitProcess = p

	var ps *os.ProcessState
	ps, err = p.Wait()

	switch {
	case err != nil:
		fmt.Fprintf(os.Stderr, "WeDeploy exit listener error: %v. Containers might still be running.\n", err)
		os.Exit(1)
	case ps.Success():
		if !dm.selfStopSignal {
			dm.waitCleanup()
		}
	default:
		println("WeDeploy exit listener failure.")
		os.Exit(1)
	}
}

func (dm *DockerMachine) maybeWaitEnd() {
	if !dm.Flags.Detach {
		go dm.waitEnd()
	}
}

func (dm *DockerMachine) waitReadyState() {
	var tries = 1
	dm.upTime = time.Now()
	dm.checkConnection()
	for tries <= 100 {
		var _, err = projects.List()

		if err == nil {
			dm.tickerd <- true
			fmt.Fprintf(dm.livew, "WeDeploy is ready! %vs\n", dm.getStartupTime())
			dm.ready()
			return
		}

		verbose.Debug(fmt.Sprintf("Trying to read projects tries #%v: %v", tries, err))
		tries++
		time.Sleep(1 * time.Second)
	}

	dm.tickerd <- true

	fmt.Fprintf(dm.livew, "WeDeploy is up.\n")
	println("Failed to verify if WeDeploy is working correctly.")
}

func (dm *DockerMachine) waitCleanup() {
	fmt.Println("WeDeploy is shutdown.")
	dm.end <- true
}

func (dm *DockerMachine) start() (err error) {
	var args = dm.getRunCommandEnv()
	var running = "docker " + strings.Join(args, " ")

	if dm.Flags.DryRun && !verbose.Enabled {
		println(running)
	} else {
		verbose.Debug(running)
	}

	if dm.Flags.DryRun {
		os.Exit(0)
	}

	if err = dm.checkPortsAreAvailable(); err != nil {
		return err
	}

	if !dm.Flags.NoUpdate && !dm.hasCurrentWeDeployImage() {
		if err := pull(); err != nil {
			return err
		}
	}

	if dm.Container, err = startCmd(args...); err != nil {
		return err
	}

	verbose.Debug("Docker container ID:", dm.Container)
	return err
}

func (dm *DockerMachine) hasCurrentWeDeployImage() bool {
	if defaults.WeDeployImageTag == dockerLatestImageTag {
		verbose.Debug("Shortcutting WeDeploy docker image as outdated (because its tag is \"latest\").")
		return false
	}

	return dm.Image == WeDeployImage
}

func (dm *DockerMachine) maybeStopListener() {
	if !dm.Flags.Detach && !dm.Flags.ViewMode {
		dm.stopListener()
	}
}

func (dm *DockerMachine) stopListener() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go dm.stopEvent(sigs)
}

func (dm *DockerMachine) stopEvent(sigs chan os.Signal) {
	<-sigs
	verbose.Debug("WeDeploy stop event called. Waiting started signal.")
	<-dm.started
	verbose.Debug("Started end signal received.")

	dm.tickerd <- true
	dm.selfStopSignal = true
	fmt.Println("\nStopping WeDeploy.")

	if err := cleanupEnvironment(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	dm.waitCleanup()
}

func (dm *DockerMachine) checkConnection() {
	var ticker = time.NewTicker(time.Second)
	go dm.checkConnectionCounter(ticker)
}

func (dm *DockerMachine) checkConnectionCounter(ticker *time.Ticker) {
	for {
		select {
		case t := <-ticker.C:
			var p = WarmupOn
			if t.Second()%2 == 0 {
				p = WarmupOff
			}

			var dots = strings.Repeat(".", t.Second()%3+1)

			fmt.Fprintf(dm.livew, "%c Starting WeDeploy%s %ds\n",
				p, dots, dm.getStartupTime())

			if err := dm.livew.Flush(); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
		case <-dm.tickerd:
			ticker.Stop()
			ticker = nil
			return
		}
	}
}

func (dm *DockerMachine) getStartupTime() int {
	return int(-dm.upTime.Sub(time.Now()).Seconds())
}

func (dm *DockerMachine) prepare() {
	dm.testAlreadyRunning()
	dm.livew = uilive.New()
	dm.tickerd = make(chan bool, 1)
	dm.started = make(chan bool, 1)
	dm.end = make(chan bool, 1)
	dm.setupPorts()
}

func (dm *DockerMachine) ready() {
	fmt.Fprintf(dm.livew, "You can now test your apps locally.")

	if !dm.Flags.ViewMode && !dm.Flags.Detach {
		fmt.Fprintf(dm.livew, " Press Ctrl+C to shut it down when you are done.")
	}

	if !dm.Flags.ViewMode && dm.Flags.Detach {
		fmt.Fprintf(dm.livew, "\nRunning on background. \"we stop\" stops the infrastructure.")
	}

	fmt.Fprintf(dm.livew, "\n")

	if err := dm.livew.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "Error flushing startup ready message: %v\n", err)
	}
}

// LoadDockerInfo loads the docker info on the DockerMachine object
func (dm *DockerMachine) LoadDockerInfo() {
	var args = []string{
		"ps",
		"--filter",
		"ancestor=" + WeDeployImage,
		"--format",
		"{{.ID}} {{.Image}}",
		"--no-trunc",
	}

	var docker = exec.Command(bin, args...)
	var buf bytes.Buffer
	docker.Stderr = os.Stderr
	docker.Stdout = &buf

	if err := docker.Run(); err != nil {
		println("docker ps error:", err.Error())
		os.Exit(1)
	}

	var ps = buf.String()

	var parts = strings.SplitAfterN(ps, " ", 2)

	switch len(parts) {
	case 0:
		dm.checkImage()
	case 2:
		dm.Container = strings.TrimSpace(parts[0])
		dm.Image = strings.TrimSpace(parts[1])
	default:
		verbose.Debug("Running docker not found on docker ps")
	}
}

func (dm *DockerMachine) checkImage() {
	var args = []string{
		"images",
		"--format",
		"{{.Repository}}:{{.Tag}}",
		"--no-trunc",
		WeDeployImage,
	}

	var docker = exec.Command(bin, args...)
	var buf bytes.Buffer
	docker.Stderr = os.Stderr
	docker.Stdout = &buf

	if err := docker.Run(); err != nil {
		println("docker images error:", err.Error())
		os.Exit(1)
	}

	dm.Image = strings.TrimSpace(buf.String())

	if dm.Image == "" {
		verbose.Debug("Docker image for the infrastructure not found.")
	}
}

func (dm *DockerMachine) testAlreadyRunning() {
	dm.LoadDockerInfo()

	// if the infrastructure is already running, test version
	if dm.Container != "" && WeDeployImage != dm.Image {
		fmt.Fprintf(os.Stderr, "docker image %v found instead of required %v\n", dm.Image, WeDeployImage)
		println("Stop the infrastructure on docker before running this command again.")
		os.Exit(1)
	}

	if !dm.Flags.DryRun && dm.Container != "" {
		verbose.Debug("Docker container ID:", dm.Container)
	}
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

func (dm *DockerMachine) getRunCommandEnv() []string {
	var address = getWeDeployHost()
	var args = []string{"run"}

	// fluentd might use either TCP or UDP, hence this special case
	args = append(args, "-p", "24224:24224/udp")

	args = append(args, dm.tcpPorts.expose()...)
	args = append(args, []string{
		"-v",
		"/var/run/docker.sock:/var/run/docker-host.sock",
		"--privileged",
		"-e",
		"WEDEPLOY_HOST_IP=" + address,
		"--detach",
		WeDeployImage,
	}...)

	return args
}

func runWait(container string) (*os.Process, error) {
	var c = exec.Command(bin, "wait", container)
	var err = c.Start()

	return c.Process, err
}

func pull() error {
	fmt.Println("Pulling WeDeploy infrastructure docker image. Hold on.")
	var docker = exec.Command(bin, "pull", WeDeployImage)
	docker.Stderr = os.Stderr
	docker.Stdout = os.Stdout

	return pullFeedback(docker.Run())
}

func pullFeedback(err error) error {
	if err == nil {
		return nil
	}

	// we ignore it for, say, "latest"
	if defaults.WeDeployImageTag != dockerLatestImageTag {
		return errwrap.Wrapf("docker pull error: {{err}}\n"+
			"Can't continue running with an outdated image", err)
	}

	return errwrap.Wrapf("docker pull error: {{err}}", err)
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
	return nil
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

	var params = []string{"rm"}
	params = append(params, ids...)
	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(params, " ")))
	var rm = exec.Command(bin, params...)
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
