// +build functional

package functional

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/api-go"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/cmdrunner"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/verbose"
)

type scenario struct {
	runCmd *cmdrunner.Command
}

var scenario1 scenario

// tests in order of execution
var tests = []func(t *testing.T){
	scenario1.setup,
	scenario1.cleanupEnvironment,
	scenario1.install,
	scenario1.testDockerStandby,
	scenario1.testDockerHasNoContainers,
	scenario1.testDockerHasNoImages,
	scenario1.testRun,
	scenario1.linkContainer,
	scenario1.testLinkedChatAfter5Seconds,
	scenario1.testErrorFeedbacks,
	scenario1.testShutdownGracefullyAfter5Seconds,
	scenario1.testDockerHasNoContainersRunning,
	scenario1.testDockerHasNoContainers,
	scenario1.teardown,
}

func TestMain(m *testing.M) {
	flag.Parse()
	validateFlags()

	print(color.Format(color.FgHiRed, "%v\n",
		"CAUTION: 100% risk of losing data if run on non-isolated staging machine."))

	if !force {
		println("Use: go test -tags=functional --channel <channel>")
		println(`Empty ("") channel runs the test with an existing "we" command.`)
		println("Skipping all functional tests.\n" +
			"Use of --force required to allow tests to run and destroy system data.")
		os.Exit(1)
	}

	ec := m.Run()
	os.Exit(ec)
}

func TestAll(t *testing.T) {
	for _, st := range tests {
		f := strings.TrimPrefix(getFunctionName(st),
			"github.com/wedeploy/cli/functional.")
		if ok := t.Run(f, st); !ok {
			break
		}
	}
}

func destroyTmp(t *testing.T) {
	if err := os.RemoveAll("tmp/"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Error trying to remove tmp/ directory: %v", err)
	}
}

func (s *scenario) setup(t *testing.T) {
	destroyTmp(t)

	if err := os.MkdirAll("tmp/", 0775); err != nil {
		t.Fatalf("Error trying to create tmp/ directory: %v", err)
	}
}

func (s *scenario) teardown(t *testing.T) {
	destroyTmp(t)
}

func getAllContainers() ([]string, error) {
	var params = []string{
		"ps", "--all", "--quiet", "--no-trunc",
	}

	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(params, " ")))
	var list = exec.Command("docker", params...)
	var buf bytes.Buffer
	list.Stderr = os.Stderr
	list.Stdout = &buf

	if err := list.Run(); err != nil {
		return []string{}, errwrap.Wrapf("Can not get containers list: {{err}}", err)
	}

	return strings.Fields(buf.String()), nil
}

func rmAllContainers() error {
	var ids, err = getAllContainers()

	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	var params = []string{"rm", "--force"}
	params = append(params, ids...)
	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(params, " ")))
	var rm = exec.Command("docker", params...)
	rm.Stderr = os.Stderr

	if err = rm.Run(); err != nil {
		return errwrap.Wrapf("Error trying to remove containers: {{err}}", err)
	}

	return err
}

func getAllImages() ([]string, error) {
	var params = []string{
		"images",
		"--all",
		"--format",
		"{{.Tag}}\t{{.ID}}",
		"--no-trunc",
	}

	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(params, " ")))
	var list = exec.Command("docker", params...)
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
		if len(i) == 2 {
			imageHashes = append(
				imageHashes,
				strings.TrimSuffix(i[1], "sha256:"))
		}
	}

	return imageHashes, nil
}

func rmAllImages() error {
	var ids, err = getAllImages()

	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return nil
	}

	var params = []string{"rmi"}
	params = append(params, ids...)
	verbose.Debug(fmt.Sprintf("Running docker %v", strings.Join(params, " ")))
	var rmi = exec.Command("docker", params...)
	rmi.Stderr = os.Stderr

	if err = rmi.Run(); err != nil {
		return errwrap.Wrapf("Error trying to remove images: {{err}}", err)
	}

	return err
}

func (s *scenario) cleanupEnvironment(t *testing.T) {
	if !cleanup {
		return
	}

	println("Running cleanup environment script")

	if err := rmAllContainers(); err != nil {
		t.Fatalf("Can not remove containers: %v", err)
	}

	if keepImages {
		log("Finished cleanup environment script (images not cleaned up)")
		return
	}

	if err := rmAllImages(); err != nil {
		t.Fatalf("Can not remove images: %v", err)
	}

	log("Finished cleanup environment script (including cleaning up images)")
}

func download() {
	var err = cmdrunner.Run("curl http://cdn.wedeploy.com/cli/latest/wedeploy.sh -sL | bash -s " + channel)

	if err != nil {
		panic(err)
	}
}

func (s *scenario) install(t *testing.T) {
	if channel != "" {
		println("Starting installing from channel " + channel)
		download()
	} else {
		log(`No --channel directive given. Using "we" in $PATH`)
	}

	if err := checkWePath(); err != nil {
		panic(errwrap.Wrapf("Path not found for we: {{err}}", err))
	}

	log("Version:")
	if err := cmdrunner.Run("we version"); err != nil {
		panic(err)
	}
}

func (s *scenario) testDockerStandby(t *testing.T) {
	var cmd = &cmdrunner.Command{
		Name: "docker",
		Args: []string{"ps", "--quiet"},
	}

	cmd.Run()

	if !cmdrunner.IsCommandOutputNop(cmd) {
		t.Errorf("Expected docker to find no containers running (command: %+v)", cmd)
	}
}

func (s *scenario) testDockerHasNoContainersRunning(t *testing.T) {
	var cmd = &cmdrunner.Command{
		Name: "docker",
		Args: []string{"ps", "--all", "--quiet"},
	}

	cmd.Run()

	if !cmdrunner.IsCommandOutputNop(cmd) {
		t.Errorf("Expected docker to find no containers running")
	}
}

func (s *scenario) testDockerHasNoContainers(t *testing.T) {
	var cmd = &cmdrunner.Command{
		Name: "docker",
		Args: []string{"ps", "--all", "--quiet"},
	}

	cmd.Run()

	if !cmdrunner.IsCommandOutputNop(cmd) {
		t.Errorf("Expected docker to find no containers")
	}
}

func (s *scenario) testDockerHasNoImages(t *testing.T) {
	if keepImages {
		t.Skipf("Images not cleaned up, jumping test")
	}

	var cmd = &cmdrunner.Command{
		Name: "docker",
		Args: []string{"images", "--quiet", "--all"},
	}

	cmd.Run()

	if !cmdrunner.IsCommandOutputNop(cmd) {
		t.Errorf("Expected docker to find no images (command: %+v)", cmd)
	}
}

var weRunFirstTimeTimeout = 15 * time.Minute

func assertReadyState(cmd *cmdrunner.Command, t *testing.T) bool {
	if cmd == nil {
		t.Fatalf(`Can not assert ready state: not invoked after "we dev --start-infra"`)
	}

	var out = cmd.Stdout.String()

	if strings.Contains(out, "Failed to verify if WeDeploy is working correctly.") {
		t.Fatalf("Unexpected infrastructure assertion error")
	}

	if !strings.Contains(out, "WeDeploy is ready!") {
		return false
	}

	u := wedeploy.URL("http://localhost/")

	if err := u.Get(); err != wedeploy.ErrUnexpectedResponse {
		t.Errorf("Expected response to be %v, got %v instead", wedeploy.ErrUnexpectedResponse, err)
	}

	log("Ready state assertion complete: infrastructure is up")
	return true
}

func (s *scenario) linkContainer(t *testing.T) {
	chdir("tmp")

	if err := cmdrunner.Run("git clone https://github.com/wedeploy/sample-wechat.git"); err != nil {
		t.Fatalf("Error trying to clone sample-wechat.git: %v", err)
	}

	chdir("sample-wechat")

	if err := cmdrunner.Run("we dev"); err != nil {
		t.Fatalf("Error trying to link sample-wechat: %v", err)
	}

	chdir("../..")
}

func (s *scenario) testLinkedChat(t *testing.T) {
	u := wedeploy.URL("http://public.whatsapp.wedeploy.me/")

	if err := apihelper.Validate(u, u.Get()); err != nil {
		t.Errorf("Expected no error for public.whatsapp.wedeploy.me page, got %v instead", err)
	} else {
		log("http://public.whatsapp.wedeploy.me/ is up")
	}

	sMsg := wedeploy.URL("http://data.whatsapp.wedeploy.me/messages")

	json := `{
	"id": "0123456789",
    "author":{
        "id":"abcdef",
        "name":"Functional test",
        "color":"color-2"
    },
    "content":"Hello, world! Message from a functional test.",
    "time":"10:00 PM"
}`

	sMsg.Body(bytes.NewBuffer([]byte(json)))

	if err := apihelper.Validate(sMsg, sMsg.Put()); err != nil {
		t.Fatalf("Error posting message to wechat: %v", err)
	}

	gMsg := wedeploy.URL("http://data.whatsapp.wedeploy.me/messages/0123456789")

	if err := gMsg.Get(); err != nil {
		t.Fatalf("Error getting posted message on wechat: %v", err)
	}

	var m map[string]interface{}
	if err := apihelper.DecodeJSON(gMsg, &m); err != nil {
		t.Fatalf("Error decoding JSON from wechat: %v", err)
	}

	v, ok := m["content"]

	if !ok || v != "Hello, world! Message from a functional test." {
		t.Fatalf("Expected value not found in %v", m)
	}
}

type feTestCase struct {
	argsList [][]string
	stderr   string
}

var feTestCases = []feTestCase{
	feTestCase{
		[][]string{
			[]string{"undeploy", "--project", "foo", "-r", "local"},
			[]string{"undeploy", "--project", "foo", "--remote", "local"},
		},
		`You can not run undeploy in the local infrastructure. Use "we dev stop" instead`,
	},
	feTestCase{
		[][]string{
			[]string{"domain", "-p", "foo", "-r", "local"},
			[]string{"domain", "--project", "foo", "-r", "local"},
			[]string{"domain", "-u", "foo.wedeploy.me"},
			[]string{"domain", "--url", "foo.wedeploy.me"},
			[]string{"domain", "-p", "foo", "-r", "local", "add", "example.com"},
			[]string{"domain", "--project", "foo", "-r", "local", "add", "example.com"},
			[]string{"domain", "-u", "foo.wedeploy.me", "-r", "local", "add", "example.com"},
			[]string{"domain", "--url", "foo.wedeploy.me", "-r", "local", "add", "example.com"},
			[]string{"domain", "-p", "foo", "-r", "local", "rm", "example.com"},
			[]string{"domain", "--project", "foo", "-r", "local", "rm", "example.com"},
			[]string{"domain", "-u", "foo.wedeploy.me", "rm", "example.com"},
			[]string{"domain", "--url", "foo.wedeploy.me", "rm", "example.com"},
			[]string{"domain", "--project", "foo", "--remote", "local"},
			[]string{"env", "-p", "foo", "-c", "bar", "-r", "local"},
			[]string{"env", "--project", "foo", "--container", "bar", "--remote", "local"},
			[]string{"env", "-u", "bar.foo.wedeploy.me"},
			[]string{"env", "--url", "bar.foo.wedeploy.me"},
			[]string{"env", "-p", "foo", "-c", "bar", "--remote", "local", "add", "envkey=envvalue"},
			[]string{"env", "--project", "foo", "-c", "bar", "add", "envkey=envvalue"},
			[]string{"env", "-u", "bar.foo.wedeploy.me", "add", "envkey=envvalue"},
			[]string{"env", "--url", "bar.foo.wedeploy.me", "add", "envkey=envvalue"},
			[]string{"env", "-p", "foo", "-c", "bar", "--remote", "local", "set", "envkey=envvalue"},
			[]string{"env", "--project", "foo", "-c", "bar", "--remote", "local", "set", "envkey=envvalue"},
			[]string{"env", "-u", "bar.foo.wedeploy.me", "set", "envkey=envvalue"},
			[]string{"env", "--url", "bar.foo.wedeploy.me", "set", "envkey=envvalue"},
			[]string{"env", "--url", "bar.foo.wedeploy.me", "set", "envkey", "envvalue"},
			[]string{"env", "-p", "foo", "-c", "bar", "--remote", "local", "rm", "envkey"},
			[]string{"env", "--project", "foo", "--container", "bar", "--remote", "local", "rm", "envkey"},
			[]string{"env", "-u", "bar.foo.wedeploy.me", "rm", "envkey"},
			[]string{"env", "--url", "bar.foo.wedeploy.me", "rm", "envkey"},
			[]string{"list", "-p", "foo", "-r", "local"},
			[]string{"list", "--project", "foo", "--remote", "local"},
			[]string{"list", "--project", "foo", "-c", "bar", "--remote", "local"},
			[]string{"list", "--project", "foo", "--container", "bar", "--remote", "local"},
			[]string{"list", "--url", "foo.wedeploy.me"},
			[]string{"log", "-p", "foo", "-r", "local"},
			[]string{"log", "--project", "foo", "--remote", "local"},
			[]string{"log", "--project", "foo", "-c", "bar", "--remote", "local"},
			[]string{"log", "--project", "foo", "--container", "bar", "--remote", "local"},
			[]string{"log", "--url", "foo.wedeploy.me"},
			[]string{"restart", "-p", "foo", "-r", "local"},
			[]string{"restart", "--project", "foo", "--remote", "local"},
			[]string{"restart", "--project", "foo", "-c", "bar", "--remote", "local"},
			[]string{"restart", "--project", "foo", "--container", "bar", "--remote", "local"},
			[]string{"restart", "--url", "foo.wedeploy.me"},
		},
		"Not found",
	},
	feTestCase{
		[][]string{
			[]string{"env", "--project", "foo", "-c", "bar", "add", "envkey"},
		},
		"Missing environment variable value",
	},
}

func testErrorFeedbackFunc(t *testing.T, args []string, stderr string) {
	var cmd = &cmdrunner.Command{
		Name: "we",
		Args: args,
	}

	cmd.Run()

	var e = &Expect{
		Stderr:   stderr,
		ExitCode: 1,
	}

	e.Assert(t, cmd)
}

func (s *scenario) testErrorFeedbacks(t *testing.T) {
	for _, tc := range feTestCases {
		for _, args := range tc.argsList {
			testErrorFeedbackFunc(t, args, tc.stderr)
		}
	}
}

func (s *scenario) testLinkedChatAfter5Seconds(t *testing.T) {
	log("Waiting 5s to test the linked chat")
	time.Sleep(5 * time.Second)
	s.testLinkedChat(t)
}

func (s *scenario) testRun(t *testing.T) {
	s.runCmd = &cmdrunner.Command{
		Name:    "we",
		Args:    []string{"dev", "--start-infra"},
		TeePipe: true,
	}

	go func() {
		log(`Executing "we dev --start-infra"`)
		s.runCmd.Start()
	}()

	var start = time.Now()
	var timeout = start.Add(weRunFirstTimeTimeout)

loop:
	var now = time.Now()

	if now.After(timeout) {
		t.Fatalf(`Timeout: %v seconds since "we dev" started.`,
			(int)(-start.Sub(time.Now()).Seconds()))
	}

	if retry := !assertReadyState(s.runCmd, t); retry {
		time.Sleep(time.Second)
		goto loop
	}
}

func (s *scenario) testShutdownGracefully(t *testing.T) {
	var ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
	var shutdown = exec.CommandContext(ctx, "we", "dev", "--shutdown-infra")

	shutdown.Stderr = os.Stderr
	shutdown.Stdout = os.Stdout

	err := shutdown.Run()
	cancel()

	if err != nil {
		t.Fatalf("we dev --shutdown-infra did not exit properly: %v.", err)
	}
}

func (s *scenario) testShutdownGracefullyAfter5Seconds(t *testing.T) {
	log("Waiting 5s to invoke we dev --shutdown-infra")
	time.Sleep(5 * time.Second)
	s.testShutdownGracefully(t)
}
