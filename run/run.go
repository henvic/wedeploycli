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

	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/projects"
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
	Detach   bool
	DryRun   bool
	ViewMode bool
	NoUpdate bool
}

// DockerMachine for the run command
type DockerMachine struct {
	Container   string
	Flags       Flags
	upTime      time.Time
	waitProcess *os.Process
	livew       *uilive.Writer
	tickerd     chan bool
	end         chan bool
	started     chan bool
}

type tcpPortsStruct []int

var tcpPorts = tcpPortsStruct{
	24224,
	80,
	5001,
	5005,
	8001,
	8080,
	8500,
	9200,
}

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
func Run(flags Flags) {
	checkDockerExists()

	var dm = &DockerMachine{
		Flags: flags,
	}

	dm.Run()
}

// Stop stops the WeDeploy infrastructure
func Stop() {
	checkDockerExists()

	var dm = &DockerMachine{}
	dm.Stop()
}

// Run executes the WeDeploy infraestruture
func (dm *DockerMachine) Run() {
	dm.prepare()

	if dm.Flags.Detach {
		dm.end <- true
	}

	if len(dm.Container) == 0 && dm.Flags.ViewMode {
		println("View mode is not available.")
		println("WeDeploy is shutdown.")
		os.Exit(1)
	}

	var already = len(dm.Container) != 0 && !dm.Flags.DryRun

	if already {
		fmt.Println("WeDeploy is already running.")
	}

	dm.maybeStopListener()

	if !already {
		dm.start()
	}

	dm.maybeWaitEnd()
	dm.started <- true
	go dm.waitReadyState()
	<-dm.end
}

// Stop stops the machine
func (dm *DockerMachine) Stop() {
	dm.testAlreadyRunning()

	if dm.Container == "" {
		println("we run is not running.")
		os.Exit(1)
	}

	stop(dm.Container)

	if dm.waitProcess != nil {
		fmt.Println("sending sigterm signal")
		dm.waitProcess.Signal(syscall.SIGTERM)
	}
}

func (dm *DockerMachine) checkPortsAreAvailable() {
	var all, notAvailable = tcpPorts.getAvailability()

	if all {
		return
	}

	println("Can't start. The following network ports must be available:")

	for _, port := range notAvailable {
		println(port)
	}

	os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "Wait call error: %v", err)
		os.Exit(1)
	case ps.Success():
		fmt.Println("WeDeploy is shutdown.")
		os.Exit(0)
	default:
		println("WeDeploy wait failure.")
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
	dm.livew.Start()
	dm.checkConnection()
	for tries <= 100 {
		var _, err = projects.List()

		if err == nil {
			dm.tickerd <- true
			time.Sleep(2 * dm.livew.RefreshInterval)
			fmt.Fprintf(dm.livew, "WeDeploy is ready!\n")
			dm.livew.Stop()
			dm.ready()
			return
		}

		verbose.Debug(fmt.Sprintf("Trying to read projects tries #%v: %v", tries, err))
		tries++
		time.Sleep(1 * time.Second)
	}

	dm.tickerd <- true

	time.Sleep(2 * dm.livew.RefreshInterval)
	fmt.Fprintf(dm.livew, "WeDeploy is up.\n")
	dm.livew.Stop()
	println("Failed to verify if WeDeploy is working correctly.")
}

func (dm *DockerMachine) start() {
	var args = getRunCommandEnv()
	var running = "docker " + strings.Join(args, " ")

	if dm.Flags.DryRun && !verbose.Enabled {
		println(running)
	} else {
		verbose.Debug(running)
	}

	if dm.Flags.DryRun {
		os.Exit(0)
	}

	dm.checkPortsAreAvailable()

	if !dm.Flags.NoUpdate || !hasCurrentWeDeployImage() {
		pull()
	}

	dm.Container = startCmd(args...)
	verbose.Debug("Docker container ID:", dm.Container)
}

func (dm *DockerMachine) stop() {
	stop(dm.Container)
	dm.end <- true
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
	fmt.Println("\nStopping WeDeploy.")
	dm.stop()
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

			fmt.Fprintf(dm.livew,
				"%c Starting WeDeploy%s %ds\n", p, dots,
				int(-dm.upTime.Sub(t).Seconds()))
		case <-dm.tickerd:
			ticker.Stop()
			ticker = nil
			return
		}
	}
}

func (dm *DockerMachine) prepare() {
	dm.upTime = time.Now()
	dm.testAlreadyRunning()
	dm.livew = uilive.New()
	dm.tickerd = make(chan bool, 1)
	dm.started = make(chan bool, 1)
	dm.end = make(chan bool, 1)
}

func (dm *DockerMachine) ready() {
	fmt.Print("You can now test your apps locally.")

	if !dm.Flags.ViewMode && !dm.Flags.Detach {
		fmt.Print(" Press Ctrl+C to shut it down when you are done.")
	}

	if !dm.Flags.ViewMode && dm.Flags.Detach {
		fmt.Print("\nRunning on background. \"we stop\" stops the infrastructure.")
	}

	fmt.Println("")
}

func (dm *DockerMachine) testAlreadyRunning() {
	var args = []string{
		"ps",
		"--filter",
		"ancestor=" + WeDeployImage,
		"--format",
		"{{.ID}}",
	}

	var docker = exec.Command(bin, args...)
	var buf bytes.Buffer
	docker.Stderr = os.Stderr
	docker.Stdout = &buf

	if err := docker.Run(); err != nil {
		println("docker ps error:", err.Error())
		os.Exit(1)
	}

	dm.Container = strings.TrimSpace(buf.String())

	if !dm.Flags.DryRun {
		verbose.Debug("Docker container ID:", dm.Container)
	}
}

func checkDockerExists() {
	if !existsDependency(bin) {
		println("Docker is not installed. Download it from http://docker.com/")
		os.Exit(1)
	}
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

func getRunCommandEnv() []string {
	var address = getWeDeployHost()
	var args = []string{"run"}

	// fluentd might use either TCP or UDP, hence this special case
	args = append(args, "-p", "24224:24224/udp")

	args = append(args, tcpPorts.expose()...)
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

func hasCurrentWeDeployImage() bool {
	if defaults.WeDeployImageTag == dockerLatestImageTag {
		verbose.Debug("Shortcutting WeDeploy docker image as outdated (because its tag is \"latest\").")
		return false
	}

	var args = []string{
		"inspect",
		"--type",
		"image",
		WeDeployImage,
	}

	var bufErrStream bytes.Buffer
	var docker = exec.Command(bin, args...)
	docker.Stderr = &bufErrStream

	err := docker.Run()

	if strings.Index(bufErrStream.String(), "Error: No such image") != -1 {
		return false
	}

	print(bufErrStream.String())

	if err != nil {
		fmt.Fprintf(os.Stderr, "docker inspect error: %v\n", err)
		return false
	}

	return true
}

func getDockerPath() string {
	var path, err = exec.LookPath(bin)

	if err != nil {
		panic(err)
	}

	return path
}

func pull() {
	fmt.Println("Pulling WeDeploy infrastructure docker image. Hold on.")
	var docker = exec.Command(bin, "pull", WeDeployImage)
	docker.Stderr = os.Stderr
	docker.Stdout = os.Stdout

	pullFeedback(docker.Run())
}

func pullFeedback(err error) {
	if err == nil {
		return
	}

	println("docker pull error:", err.Error())

	if defaults.WeDeployImageTag != dockerLatestImageTag {
		os.Exit(1)
	}
}

func startCmd(args ...string) string {
	verbose.Debug("Starting WeDeploy")
	var docker = exec.Command(bin, args...)
	var dockerContainerBuf bytes.Buffer
	docker.Stderr = os.Stderr
	docker.Stdout = &dockerContainerBuf

	if err := docker.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "docker run error:", err)
		os.Exit(1)
	}

	var dockerContainer = strings.TrimSpace(dockerContainerBuf.String())
	return dockerContainer
}

func stop(container string) {
	var stop = exec.Command(bin, "stop", container)

	stopFeedback(stop.Run())
}

func stopFeedback(err error) {
	switch err.(type) {
	case nil:
	case *exec.ExitError:
		println("warning: still stopping WeDeploy on background\n")
		os.Exit(1)
	default:
		println("docker stop error:", err.Error())
		os.Exit(1)
	}
}

func existsDependency(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
