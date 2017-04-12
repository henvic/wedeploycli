package run

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/exechelper"
	"github.com/wedeploy/cli/status"
	"github.com/wedeploy/cli/user"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/waitlivemsg"
)

// Flags modifiers
type Flags struct {
	Debug  bool
	DryRun bool
}

// DockerMachine for the run command
type DockerMachine struct {
	Container      string
	Image          string
	Flags          Flags
	WaitLiveMsg    *waitlivemsg.WaitLiveMsg
	livew          *uilive.Writer
	end            chan bool
	started        chan bool
	Context        context.Context
	contextCancel  context.CancelFunc
	selfStopSignal bool
	tcpPorts       tcpPortsStruct
}

// Run runs the WeDeploy infrastructure
func Run(ctx context.Context, flags Flags) error {
	if err := checkDockerAvailable(); err != nil {
		return err
	}

	var dm = &DockerMachine{
		Flags:   flags,
		Context: ctx,
	}

	return dm.Run()
}

// Stop stops the WeDeploy infrastructure
func Stop() error {
	if err := checkDockerAvailable(); err != nil {
		return err
	}

	var dm = &DockerMachine{
		Context: context.Background(),
	}
	return dm.Stop()
}

func (dm *DockerMachine) checkDockerDebug() (ok bool, err error) {
	var docker = exec.CommandContext(dm.Context, bin, "port", dm.Container)
	exechelper.AddCommandToNewProcessGroup(docker)
	docker.Stderr = os.Stderr
	var buf bytes.Buffer
	docker.Stdout = &buf
	err = docker.Run()
	var s = buf.String()
	ok = strings.Contains(s,
		fmt.Sprintf("%v/tcp", debugPorts[0]))

	return ok, err
}

// Run executes the WeDeploy infrastruture
func (dm *DockerMachine) Run() (err error) {
	dm.Context, dm.contextCancel = context.WithCancel(dm.Context)

	if err = dm.LoadDockerInfo(); err != nil {
		return err
	}

	if dm.Container == "" {
		switch dashboardID, dashboardOnHostErr := dm.checkDashboardIsOnOnHost(); {
		case dashboardOnHostErr != nil:
			return err
		case dashboardID != "":
			println(color.Format(color.FgHiYellow, "Developer") + " infrastructure is on.")
			return nil
		default:
			verbose.Debug("No developer infrastructure found.")
		}
	}

	dm.setupPorts()

	if !dm.Flags.DryRun && dm.Container != "" {
		verbose.Debug(`Infrastructure is on.`)

		if dm.Flags.Debug {
			var ok, errd = dm.checkDockerDebug()

			if errd != nil {
				return errwrap.Wrapf("Can not get docker ports for container: {{err}}", errd)
			}

			if !ok {
				fmt.Fprintf(os.Stderr, "%v\n",
					color.Format(color.BgRed, color.Bold,
						" change to debug mode not allowed: already running infrastructure "))

				fmt.Fprintf(os.Stderr, "%v\n", `
To run the infrastructure with debug mode:
	1. Shutdown with "we run --shutdown-infra"
	2. Run with "we run --infra --debug"
	3. Run any project or containers you want`)
			}
		}

		return nil
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

	dm.started <- true
	go dm.dockerWait()
	go dm.waitReadyState()
	<-dm.end
	return dm.createUser()
}

func (dm *DockerMachine) createUser() (err error) {
	_, err = user.Create(context.Background(), &user.User{
		Email:    "no-reply@wedeploy.com",
		Password: "cli-tool-password",
		Name:     "CLI Tool",
	})

	if err != nil {
		return errwrap.Wrapf("Failed to authenticate: {{err}}", err)
	}

	fmt.Fprintf(dm.livew, "WeDeploy is ready! %vs\n", dm.WaitLiveMsg.Duration())
	dm.WaitLiveMsg.Stop()
	_ = dm.livew.Flush()

	return err
}

func (dm *DockerMachine) dockerWait() {
	var docker = exec.CommandContext(dm.Context, bin, "wait", dm.Container)
	exechelper.AddCommandToNewProcessGroup(docker)
	_ = docker.Run()
	dm.contextCancel()
	fmt.Fprintf(os.Stderr, "Infrastructure terminated unexpectedly.\n")
	dm.end <- true
}

// Stop stops the machine
func (dm *DockerMachine) Stop() error {
	dm.livew = uilive.New()
	stopMsg := waitlivemsg.WaitLiveMsg{
		Msg:    "Stopping WeDeploy.",
		Stream: dm.livew,
	}

	go stopMsg.Wait()

	if err := dm.LoadDockerInfo(); err != nil {
		return err
	}

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

	stopMsg.Stop()
	fmt.Println("WeDeploy is shutdown.")

	return nil
}

var servicesPorts = tcpPortsStruct{
	80,
	3002,

	5001,
	5005,
	8500,
	9200,

	6379,
	8001,
	8080,
	8081,
	8082,
	24224,
}

var debugPorts = tcpPortsStruct{}

func (dm *DockerMachine) setupPorts() {
	dm.tcpPorts = servicesPorts

	if dm.Flags.Debug {
		dm.tcpPorts = append(dm.tcpPorts, debugPorts...)
	}
}

func (dm *DockerMachine) checkPortsAreAvailable() error {
	var all, notAvailable = dm.tcpPorts.getAvailability()

	if all {
		return nil
	}

	var s = "Can not start. The following network ports must be available:\n"

	for _, port := range notAvailable {
		s += fmt.Sprintf("%v\n", port)
	}

	s += "\nSometimes docker doesn't free up ports properly.\n" +
		"If you don't know why these ports are not available, try restarting docker."

	return errors.New(s)
}

func (dm *DockerMachine) waitReadyState() {
	fmt.Println("WeDeploy is not running yet... Please wait.")
	var tries = 1
	dm.WaitLiveMsg = &waitlivemsg.WaitLiveMsg{
		Msg:    "Starting WeDeploy",
		Stream: dm.livew,
	}

	go dm.WaitLiveMsg.Wait()

	// Starting WeDeploy
	for tries <= 100 || dm.WaitLiveMsg.Duration() < 300 {
		verbose.Debug(fmt.Sprintf("Trying #%v", tries))
		tries++
		var ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		var status, err = status.Get(ctx)
		cancel()

		if err != nil || status.Status != "up" {
			verbose.Debug("System not available:", status, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		dm.end <- true
	}
}

func (dm *DockerMachine) endRun() {
	fmt.Println("WeDeploy is shutdown.")
	dm.end <- true
}

func (dm *DockerMachine) start() (err error) {
	if err = dm.maybeInitSwarm(); err != nil {
		return err
	}

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

	if err := dm.checkDockerHost(); err != nil {
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

func (dm *DockerMachine) maybeInitSwarm() error {
	var params = []string{
		"swarm", "init",
	}

	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(params, " ")))
	var swarmErrBuf bytes.Buffer
	var swarm = exec.Command(bin, params...)
	swarm.Stderr = &swarmErrBuf
	var err = swarm.Run()

	if strings.Contains(swarmErrBuf.String(), "This node is already part of a swarm.") {
		verbose.Debug("Skipping swarm initialization: host is already part of a swarm")
		return nil
	}

	if swarmErrBuf.Len() != 0 {
		fmt.Fprintf(os.Stderr, "%v", swarmErrBuf.String())
	}

	return err
}

func (dm *DockerMachine) checkDockerHost() error {
	dh, ok := os.LookupEnv("DOCKER_HOST")

	if !ok {
		return nil
	}

	if _, err := os.Stat(dh); err == nil || !os.IsNotExist(err) {
		return nil
	}

	var m = `Can not work with $DOCKER_HOST env variable set to non-socket.`

	if runtime.GOOS != "linux" {
		m += `If you are using docker-machine, please use Docker Native instead.
Download it from http://docker.com/`
	}

	return errors.New(m)
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
			println("\n\"we run\" killed awkwardly. Use \"we run --shutdown-infra\" to kill ghosts.")
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

func (dm *DockerMachine) checkDashboardIsOnOnHost() (string, error) {
	var args = []string{"ps", "--filter", "ancestor=wedeploy/dashboard", "--quiet"}

	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(args, " ")))
	var docker = exec.CommandContext(dm.Context, bin, args...)
	exechelper.AddCommandToNewProcessGroup(docker)
	var buf bytes.Buffer
	docker.Stderr = os.Stderr
	docker.Stdout = &buf

	if err := docker.Run(); err != nil {
		return "", errwrap.Wrapf("docker ps error: {{err}}", err)
	}

	return strings.TrimSpace(buf.String()), nil
}

// LoadDockerInfo loads the docker info on the DockerMachine object
func (dm *DockerMachine) LoadDockerInfo() error {
	var args = []string{
		"ps",
		"--filter",
		"ancestor=" + WeDeployImage,
		"--format",
		"{{.ID}} {{.Image}}",
		"--no-trunc",
	}

	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(args, " ")))
	var docker = exec.CommandContext(dm.Context, bin, args...)
	exechelper.AddCommandToNewProcessGroup(docker)
	var buf bytes.Buffer
	docker.Stderr = os.Stderr
	docker.Stdout = &buf

	if err := docker.Run(); err != nil {
		return errwrap.Wrapf("docker ps error: {{err}}", err)
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

	return nil
}

func (dm *DockerMachine) checkImage() {
	var args = []string{
		"images",
		"--format",
		"{{.Repository}}:{{.Tag}}",
		"--no-trunc",
		WeDeployImage,
	}

	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(args, " ")))
	var docker = exec.CommandContext(dm.Context, bin, args...)
	exechelper.AddCommandToNewProcessGroup(docker)
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
	_ = getWeDeployHost()
	var args = []string{"run"}

	args = append(args, dm.tcpPorts.expose()...)
	args = append(args, []string{
		"--volume",
		"/var/run/docker.sock:/var/run/docker.sock",
		"--network",
		DockerNetwork,
		"--name",
		"wedeploy-local",
		"--network-alias",
		"api",
		"--network-alias",
		"auth",
		"--network-alias",
		"data",
		"--network-alias",
		"email",
		"--network-alias",
		"wedeploy-conqueror",
		"--network-alias",
		"wedeploy-consul-server",
		"--network-alias",
		"wedeploy-elasticsearch-log",
		"--network-alias",
		"wedeploy-elasticsearch",
		"--network-alias",
		"wedeploy-redis-infrastructure",
		"--label",
		"com.wedeploy.local=true",
		"--detach",
		WeDeployImage,
	}...)

	return args
}
