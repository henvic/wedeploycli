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

// Flags modifiers
type Flags struct {
	Detach   bool
	DryRun   bool
	ViewMode bool
}

// DockerMachine for the run command
type DockerMachine struct {
	Container string
	Flags     Flags
	upTime    time.Time
	livew     *uilive.Writer
	tickerd   chan bool
	end       chan bool
	started   chan bool
}

var portsArgs = []string{
	"-p", "53:53/tcp",
	"-p", "53:53/udp",
	"-p", "80:80",
	"-p", "5001:5001",
	"-p", "5005:5005",
	"-p", "8001:8001",
	"-p", "8080:8080",
	"-p", "8500:8500",
	"-p", "9200:9200",
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
}

func (dm *DockerMachine) waitEnd() {
	var ps, err = waitEnd(dm.Container)

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

	if !hasCurrentWeDeployImage() {
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
				"%c connecting%s %ds\n", p, dots,
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
	verbose.Debug("Docker container ID:", dm.Container)
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

	args = append(args, portsArgs...)
	args = append(args, []string{
		"--privileged",
		"-e",
		"WEDEPLOY_HOST_IP=" + address,
		"--detach",
		WeDeployImage,
	}...)

	return args
}

func hasCurrentWeDeployImage() bool {
	var args = []string{
		"inspect",
		"--type",
		"image",
		WeDeployImage,
	}

	var docker = exec.Command(bin, args...)
	docker.Stderr = os.Stderr

	if err := docker.Run(); err != nil {
		verbose.Debug("docker inspect error:", err.Error())
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

	if err := docker.Run(); err != nil {
		println("docker pull error:", err.Error())
		os.Exit(1)
	}
}

func startCmd(args ...string) string {
	fmt.Println("Starting WeDeploy")
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

	if err := stop.Run(); err != nil {
		println("docker stop error:", err.Error())
		os.Exit(1)
	}
}

func existsDependency(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func waitEnd(container string) (*os.ProcessState, error) {
	var procWait, err = os.StartProcess(getDockerPath(),
		[]string{bin, "wait", container},
		&os.ProcAttr{
			Sys: &syscall.SysProcAttr{
				Setpgid: true,
			},
			Files: []*os.File{nil, nil, nil},
		})

	if err != nil {
		return nil, err
	}

	return procWait.Wait()
}
