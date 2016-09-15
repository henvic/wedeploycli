package run

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gosuri/uilive"
	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/verbose"
)

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

	if err := unlinkProjects(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
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

	if _, err = p.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "WeDeploy exit listener error: %v. Containers might still be running.\n", err)
		os.Exit(1)
	}

	if !dm.selfStopSignal {
		dm.waitCleanup()
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

	var killListenerStarted sync.WaitGroup
	killListenerStarted.Add(1)

	go func() {
		killListenerStarted.Done()
		<-sigs
		println("Cleaning up running infrastructure. Please wait.")
		<-sigs
		println("To kill this window (not recommended), try again in 60 seconds.")
		var gracefulExitLoopTimeout = time.Now().Add(1 * time.Minute)
	killLoop:
		<-sigs

		if time.Now().After(gracefulExitLoopTimeout) {
			println("\n\"we run\" killed awkwardly. Use \"we stop\" to kill ghosts.")
			os.Exit(1)
		}

		goto killLoop
	}()

	killListenerStarted.Wait()

	if err := unlinkProjects(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

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
	tryAddCommandToNewProcessGroup(docker)
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
	tryAddCommandToNewProcessGroup(docker)
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
