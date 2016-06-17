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
	"sync"
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

// Flags modifiers
type Flags struct {
	Detach   bool
	DryRun   bool
	ViewMode bool
}

var portsArgs = []string{
	"-p", "53:53/tcp",
	"-p", "53:53/udp",
	"-p", "80:80",
	"-p", "8001:8001",
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
	if !existsDependency("docker") {
		println("Docker is not installed. Download it from http://docker.com/")
		os.Exit(1)
	}

	var dockerContainer = getAlreadyRunning()

	if len(dockerContainer) != 0 && !flags.DryRun {
		fmt.Println("WeDeploy is already running.")
	} else if !flags.ViewMode {
		dockerContainer = start(flags)
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			waitReadyState()
			wg.Done()
		}()
		wg.Wait()
	} else {
		println("View mode is not available.")
		println("WeDeploy is shutdown.")
		os.Exit(1)
	}

	verbose.Debug("Docker container ID:", dockerContainer)

	if !flags.Detach && !flags.ViewMode {
		stopListener(dockerContainer)
	}

	fmt.Print("You can now test your apps locally.")

	if !flags.ViewMode && !flags.Detach {
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
		"ancestor=" + WeDeployImage,
		"--format",
		"{{.ID}}",
	}

	var docker = exec.Command("docker", args...)

	var dockerContainerBuf bytes.Buffer
	docker.Stderr = os.Stderr
	docker.Stdout = &dockerContainerBuf

	if err := docker.Run(); err != nil {
		println("docker ps error:", err.Error())
		os.Exit(1)
	}

	var dockerContainer = strings.TrimSpace(dockerContainerBuf.String())
	return dockerContainer
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

	var docker = exec.Command("docker", args...)
	docker.Stderr = os.Stderr

	if err := docker.Run(); err != nil {
		verbose.Debug("docker inspect error:", err.Error())
		return false
	}

	return true
}

func getDockerPath() string {
	var path, err = exec.LookPath("docker")

	if err != nil {
		panic(err)
	}

	return path
}

func dockerWait(dockerContainer string) *os.Process {
	var procWait, err = os.StartProcess(getDockerPath(),
		[]string{"docker", "wait", dockerContainer},
		&os.ProcAttr{
			Sys: &syscall.SysProcAttr{
				Setpgid: true,
			},
			Files: []*os.File{nil, nil, nil},
		})

	if err != nil {
		println("WeDeploy wait error:", err.Error())
		os.Exit(1)
	}

	return procWait
}

func listen(dockerContainer string) {
	var ps, err = dockerWait(dockerContainer).Wait()

	if err != nil {
		println("WeDeploy wait.Wait error:", err.Error())
		os.Exit(1)
	}

	if ps.Success() {
		fmt.Println("WeDeploy is shutdown.")
	} else {
		println("WeDeploy wait failure.")
	}
}

func pull() {
	fmt.Println("Pulling WeDeploy infrastructure docker image. Hold on.")
	var docker = exec.Command("docker", "pull", WeDeployImage)
	docker.Stderr = os.Stderr
	docker.Stdout = os.Stdout

	if err := docker.Run(); err != nil {
		println("docker pull error:", err.Error())
		os.Exit(1)
	}
}

func waitReadyState() {
	var tries = 1
	var livew = uilive.New()
	verbose.Debug("Waiting 10 seconds for ready state")
	livew.Start()
	var tdone = warmup(livew)
	time.Sleep(10 * time.Second)
	for tries <= 20 {
		var _, err = projects.List()

		if err == nil {
			close(tdone)
			time.Sleep(2 * livew.RefreshInterval)
			fmt.Fprintf(livew, "WeDeploy is ready!\n")
			livew.Stop()
			return
		}

		verbose.Debug(fmt.Sprintf("Trying to read projects tries #%v: %v", tries, err))
		tries++
		time.Sleep(4 * time.Second)
	}

	close(tdone)
	time.Sleep(2 * livew.RefreshInterval)
	fmt.Fprintf(livew, "WeDeploy is up.\n")
	livew.Stop()
	println("WeDeploy is online, but failed to read projects to verify everything is fine.")
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

	if !hasCurrentWeDeployImage() {
		pull()
	}

	return startCmd(args...)
}

func startCmd(args ...string) string {
	fmt.Println("Starting WeDeploy")
	var docker = exec.Command("docker", args...)
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
		fmt.Println("\nStopping WeDeploy.")
		stop(dockerContainer)
	}()
}

func existsDependency(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func warmup(writer *uilive.Writer) chan bool {
	var ticker = time.NewTicker(time.Second)
	var tdone = make(chan bool, 1)

	go warmupCounter(writer, ticker, tdone)

	return tdone
}

func warmupCounter(w *uilive.Writer, ticker *time.Ticker, tdone chan bool) {
	var now = time.Now()

	for {
		select {
		case t := <-ticker.C:
			var p = WarmupOn
			if t.Second()%2 == 0 {
				p = WarmupOff
			}

			fmt.Fprintf(w, "%c warming up %ds\n", p, int(-now.Sub(t).Seconds()))
		case <-tdone:
			ticker.Stop()
			ticker = nil
			return
		}
	}
}
