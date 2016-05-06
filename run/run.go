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

	"github.com/launchpad-project/cli/verbose"
)

// ErrHostNotFound is used when host is not found
var ErrHostNotFound = errors.New("You need to be connected to a network.")

// Flags modifiers
type Flags struct {
	Detach   bool
	DryRun   bool
	ViewMode bool
}

var portsArgs = []string{
	"-p", "80:80",
	"-p", "9300:9300",
	"-p", "5701:5701",
	"-p", "8001:8001",
	"-p", "8080:8080",
	"-p", "5005:5005",
}

// GetLaunchpadHost gets the Launchpad infrastructure host
// This is a temporary solution and it is NOT reliable
func GetLaunchpadHost() (string, error) {
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

// Run runs the Launchpad infrastructure
func Run(flags Flags) {
	if !existsDependency("docker") {
		println("Docker is not installed. Download it from http://docker.com/")
		os.Exit(1)
	}

	var dockerContainer = getAlreadyRunning()

	if len(dockerContainer) != 0 {
		fmt.Println("Launchpad is already running.")
	} else if !flags.ViewMode {
		dockerContainer = start(flags)
	} else {
		println("View mode is not available.")
		println("Launchpad is shutdown.")
		os.Exit(1)
	}

	verbose.Debug("Docker container ID:", dockerContainer)

	if !flags.Detach && !flags.ViewMode {
		stopListener(dockerContainer)
	}

	fmt.Print("You can now test your apps locally.")

	if !flags.ViewMode {
		fmt.Print(" Press Ctrl+C to shut it down when you are done.")
	}

	fmt.Println("")

	if !flags.Detach {
		listen(dockerContainer)
	}
}

func getAlreadyRunning() string {
	var args = []string{
		"ps",
		"--filter",
		"ancestor=launchpad/dev",
		"--format",
		"{{.ID}}",
	}

	var docker = exec.Command("docker", args...)

	var dockerContainerBuf bytes.Buffer
	docker.Stdout = &dockerContainerBuf

	if err := docker.Run(); err != nil {
		println("docker ps error:", err.Error())
		os.Exit(1)
	}

	var dockerContainer = strings.TrimSpace(dockerContainerBuf.String())
	return dockerContainer
}

func getRunCommandEnv() []string {
	var args = []string{"run"}
	args = append(args, portsArgs...)

	var address, err = GetLaunchpadHost()

	if err != nil {
		panic(err)
	}

	args = append(args, "-e")
	args = append(args, "LP_DEV_IP_ADDRESS="+address)
	args = append(args, "-e")
	args = append(args, "LP_DEV_DOCKER_HOST=tcp://"+address+":2375")
	args = append(args, "--detach")
	args = append(args, "launchpad/dev")
	return args
}

func listen(dockerContainer string) {
	dockerPath, err := exec.LookPath("docker")

	if err != nil {
		panic(err)
	}

	procWait, err := os.StartProcess(dockerPath,
		[]string{"docker", "wait", dockerContainer},
		&os.ProcAttr{
			Sys: &syscall.SysProcAttr{
				Setpgid: true,
			},
			Files: []*os.File{nil, nil, nil},
		})

	if err != nil {
		println("Launchpad wait error:", err.Error())
		os.Exit(1)
	}

	ps, err := procWait.Wait()

	if err != nil {
		println("Launchpad wait.Wait error:", err.Error())
		os.Exit(1)
	}

	if ps.Success() {
		fmt.Println("Launchpad is shutdown.")
	} else {
		println("Launchpad wait failure.")
	}
}

func start(flags Flags) string {
	var args = getRunCommandEnv()
	var running = "docker " + strings.Join(args, " ")

	if flags.DryRun && !verbose.Enabled {
		println(running)
	} else {
		verbose.Debug(running)
	}

	if flags.DryRun {
		os.Exit(0)
	}

	return startCmd(args...)
}

func startCmd(args ...string) string {
	fmt.Println("Starting Launchpad")
	var docker = exec.Command("docker", args...)
	var dockerContainerBuf bytes.Buffer
	docker.Stdout = &dockerContainerBuf

	if err := docker.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "docker run error:", err)
		os.Exit(1)
	}

	var dockerContainer = strings.TrimSpace(dockerContainerBuf.String())
	return dockerContainer
}

func stop(dockerContainer string) {
	var stop = exec.Command("docker", "stop", dockerContainer)

	if err := stop.Run(); err != nil {
		println("docker stop error:", err.Error())
		os.Exit(1)
	}
}

func stopListener(dockerContainer string) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("\nStopping Launchpad.")
		stop(dockerContainer)
	}()
}

func existsDependency(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
