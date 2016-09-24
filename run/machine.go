package run

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/waitlivemsg"
)

// Flags modifiers
type Flags struct {
	Debug  bool
	Detach bool
	DryRun bool
}

// DockerMachine for the run command
type DockerMachine struct {
	Container   string
	Image       string
	Flags       Flags
	WaitLiveMsg *waitlivemsg.WaitLiveMsg
	// waitProcess is the "docker wait" PID
	waitProcess    *os.Process
	livew          *uilive.Writer
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
	dm.LoadDockerInfo()
	dm.setupPorts()

	if !dm.Flags.DryRun && dm.Container != "" {
		println(`Infrastructure already running.`)
		println(`Use "we stop" to stop it.`)
		os.Exit(0)
	}

	dm.livew = uilive.New()
	dm.started = make(chan bool, 1)
	dm.end = make(chan bool, 1)

	if !dm.Flags.DryRun {
		if err = cleanupEnvironment(); err != nil {
			return err
		}
	}

	dm.stopListener()

	if err = dm.start(); err != nil {
		return err
	}

	dm.maybeWaitEnd()
	dm.started <- true
	go dm.waitReadyState()
	<-dm.end

	return nil
}

// Stop stops the machine
func (dm *DockerMachine) Stop() error {
	dm.livew = uilive.New()
	stopMsg := waitlivemsg.WaitLiveMsg{
		Msg:    "Stopping WeDeploy.",
		Stream: dm.livew,
	}

	go stopMsg.Wait()
	dm.LoadDockerInfo()

	if dm.Container == "" {
		verbose.Debug("No infrastructure container detected.")
	}

	if err := unlinkProjects(); err != nil &&
		!strings.Contains(err.Error(), "local infrastructure is not running") {
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

	stopMsg.Stop()
	fmt.Println("WeDeploy is shutdown.")

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

	var s = "Can't start. The following network ports must be available:\n"

	for _, port := range notAvailable {
		s += fmt.Sprintf("%v\n", port)
	}

	s += "\nSometimes docker doesn't free up ports properly.\n" +
		"If you don't know why these ports are not available, try restarting docker."

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
		dm.endRun()
	}
}

func (dm *DockerMachine) maybeWaitEnd() {
	if !dm.Flags.Detach {
		go dm.waitEnd()
	}
}

func (dm *DockerMachine) waitReadyState() {
	var tries = 1
	dm.WaitLiveMsg = &waitlivemsg.WaitLiveMsg{
		Msg:    "Starting WeDeploy",
		Stream: dm.livew,
	}

	go dm.WaitLiveMsg.Wait()

	// Starting WeDeploy
	for tries <= 100 {
		var ctx, cancel = context.WithTimeout(context.Background(), 500*time.Millisecond)
		var _, err = projects.List(ctx)

		cancel()

		if err == nil {
			dm.WaitLiveMsg.Stop()
			fmt.Fprintf(dm.livew, "WeDeploy is ready! %vs\n", dm.WaitLiveMsg.Duration())
			dm.ready()
			if dm.Flags.Detach {
				dm.end <- true
			}
			return
		}

		verbose.Debug(fmt.Sprintf("Trying to read projects tries #%v: %v", tries, err))
		tries++
		time.Sleep(1 * time.Second)
	}

	dm.WaitLiveMsg.Stop()

	println("Failed to verify if WeDeploy is working correctly.")
	if dm.Flags.Detach {
		dm.end <- true
	}
}

func (dm *DockerMachine) endRun() {
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

	if err = dm.checkPortsAreAvailable(); err != nil {
		return err
	}

	if dm.Flags.DryRun {
		os.Exit(0)
	}

	if dm.Container, err = startCmd(args...); err != nil {
		return err
	}

	verbose.Debug("Docker container ID:", dm.Container)
	return err
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

	dm.WaitLiveMsg.Stop()
	dm.selfStopSignal = true
	fmt.Println("")

	stopMsg := waitlivemsg.WaitLiveMsg{
		Msg:    "Stopping WeDeploy.",
		Stream: dm.livew,
	}

	go stopMsg.Wait()

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

	stopMsg.Stop()
	dm.endRun()
}

func (dm *DockerMachine) ready() {
	fmt.Fprintf(dm.livew, "You can now test your apps locally.")

	if !dm.Flags.Detach {
		fmt.Fprintf(dm.livew, " Press Ctrl+C to shut it down when you are done.")
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
