package run

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"sync"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/createuser"
	"github.com/wedeploy/cli/exechelper"
	"github.com/wedeploy/cli/status"
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
	waitLiveMsg    waitlivemsg.WaitLiveMsg
	livew          *uilive.Writer
	end            chan bool
	started        chan bool
	Context        context.Context
	contextCancel  context.CancelFunc
	tcpPorts       tcpPortsMap
	terminate      bool
	terminateMutex sync.RWMutex
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
	1. Shutdown with "we deploy --stop-local-infra"
	2. Run with "we deploy --infra --debug"
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

	dm.beginStopListener()

	if err = dm.start(); err != nil {
		return err
	}

	dm.started <- true
	go dm.dockerWait()
	go dm.waitReadyState()
	<-dm.end

	defer dm.terminateMutex.Unlock()
	dm.terminateMutex.Lock()
	if dm.terminate {
		return nil
	}

	return dm.createUser()
}

func (dm *DockerMachine) createUser() (err error) {
	if err := createuser.Try(dm.Context); err != nil {
		return err
	}

	dm.waitLiveMsg.StopWithMessage(fmt.Sprintf("WeDeploy is ready! %vs", dm.waitLiveMsg.Duration()))
	_ = dm.livew.Flush()

	return err
}

func (dm *DockerMachine) dockerWait() {
	var docker = exec.CommandContext(dm.Context, bin, "wait", dm.Container)
	exechelper.AddCommandToNewProcessGroup(docker)
	_ = docker.Run()
	dm.contextCancel()

	if !dm.terminating() {
		fmt.Fprintf(os.Stderr, "Infrastructure terminated unexpectedly.\n")
	}

	dm.end <- true
}

var servicesPorts = tcpPortsMap{
	TCPPort{
		Internal: 5001,
		Expose:   5001,
	},
	TCPPort{
		Internal: 5005,
		Expose:   5005,
	},
	TCPPort{
		Internal: 8500,
		Expose:   8500,
	},
	TCPPort{
		Internal: 9200,
		Expose:   9200,
	},

	TCPPort{
		Internal: 6379,
		Expose:   6379,
	},
	TCPPort{
		Internal: 8001,
		Expose:   8001,
	},
	TCPPort{
		Internal: 8080,
		Expose:   8080,
	},
	TCPPort{
		Internal: 8081,
		Expose:   8081,
	},
	TCPPort{
		Internal: 8082,
		Expose:   8082,
	},
	TCPPort{
		Internal: 24224,
		Expose:   24224,
	},
}

var debugPorts = tcpPortsMap{
	TCPPort{
		Internal: 3002,
		Expose:   3002,
	},
}

func (dm *DockerMachine) setupPorts() {
	dm.tcpPorts = servicesPorts

	dm.tcpPorts = append(dm.tcpPorts, TCPPort{
		Internal: 80,
		Expose:   config.Global.LocalHTTPPort,
	})

	dm.tcpPorts = append(dm.tcpPorts, TCPPort{
		Internal: 443,
		Expose:   config.Global.LocalHTTPSPort,
	})

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
	var tries = 1

	dm.waitLiveMsg.SetStream(dm.livew)
	dm.waitLiveMsg.SetMessage("WeDeploy is starting")
	go dm.waitLiveMsg.Wait()

	// Starting WeDeploy
	for tries <= 100 || dm.waitLiveMsg.Duration() < 300 {
		verbose.Debug(fmt.Sprintf("Trying #%v", tries))
		tries++
		var ctx, cancel = context.WithTimeout(dm.Context, time.Second)
		var isUp = status.IsUp(ctx)
		cancel()

		if !isUp {
			verbose.Debug("System not available")
			time.Sleep(100 * time.Millisecond)
			continue
		}

		dm.end <- true
	}
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
