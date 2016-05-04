package run

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/launchpad-project/cli/verbose"
)

// ErrHostNotFound is used when host is not found
var ErrHostNotFound = errors.New("Host not found")

// ErrInvalidHost is used when host is in invalid format
type ErrInvalidHost struct {
	Reference error
}

func (e ErrInvalidHost) Error() string {
	return fmt.Sprintf("Host is invalid: %v", e.Reference)
}

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
func GetLaunchpadHost() (string, error) {
	var fromEnv, err = getLaunchpadHostFromEnv()

	if err == nil {
		return fromEnv, nil
	}

	if err != ErrHostNotFound {
		return "", ErrInvalidHost{
			Reference: err,
		}
	}

	verbose.Debug("Environment variable $DOCKER_HOST not found.")
	verbose.Debug("Falling back to localhost")
	return "localhost", nil
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
	var args = []string{"run"}
	var certPath, hasCertPath = os.LookupEnv("DOCKER_CERT_PATH")
	args = append(args, portsArgs...)

	if hasCertPath {
		args = append(args, "-v")
		args = append(args, certPath+":/certs")
	} else {
		verbose.Debug("Environment var $DOCKER_CERT_PATH not found. Ignoring.")
	}

	var address, err = GetLaunchpadHost()

	if err != nil {
		panic(err)
	}

	args = append(args, "-e")
	args = append(args, "LP_DEV_IP_ADDRESS="+address)
	args = append(args, "--detach")
	args = append(args, "launchpad/dev")

	var running = "docker " + strings.Join(args, " ")

	if flags.DryRun && !verbose.Enabled {
		println(running)
	} else {
		verbose.Debug(running)
	}

	fmt.Println("Starting Launchpad")

	var docker = exec.Command("docker", args...)
	var dockerContainerBuf bytes.Buffer
	docker.Stdout = &dockerContainerBuf

	if err := docker.Run(); err != nil {
		println("docker run error:", err.Error())
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

func getLaunchpadHostFromEnv() (string, error) {
	var dhost, exists = os.LookupEnv("DOCKER_HOST")

	if !exists {
		return "", ErrHostNotFound
	}

	var u, err = url.Parse(dhost)

	if err != nil {
		return "", err
	}

	host, _, err := net.SplitHostPort(u.Host)
	return host, err
}

func existsDependency(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
