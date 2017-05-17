package functional

// no http handlers will be used for functional tests
// nothing should be mocked for them

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"runtime"
	"testing"
	"time"

	"github.com/wedeploy/cli/cmdrunner"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/stringlib"
	"github.com/wedeploy/cli/verbosereq"
)

var (
	channel    = ""
	force      = false
	cleanup    = false
	keepImages = true

	logTimeFormat = "Jan _2 06 15:04:05 MST"
)

func init() {
	flag.StringVar(&channel, "channel", "", "distribution channel to test (empty for jumping download)")
	flag.BoolVar(&force, "force", false, "force running the tests")
	flag.BoolVar(&cleanup, "cleanup", false, "clean up environment (docker, etc) before running tests")
	flag.BoolVar(&keepImages, "keep-images", true, "keep images on cleanup")
	verbosereq.SetLogFunc(logInterface)
}

func chdir(dir string) {
	log("chdir " + dir)
	if ech := os.Chdir(dir); ech != nil {
		panic(ech)
	}
}

func checkWePath() error {
	var path, err = exec.LookPath("we")

	if err == nil {
		fmt.Println("\nPath: " + path)
	}

	return err
}

func validateChannel() {
	re := regexp.MustCompile("^$|^[a-zA-Z0-9]+$")

	if !re.MatchString(channel) {
		panic(errors.New("invalid channel name"))
	}
}

func validateFlags() {
	validateChannel()
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func log(format string, a ...interface{}) {
	logTimeFormat := color.Format(color.FgHiRed, "["+time.Now().Format(logTimeFormat)+"]")
	format = logTimeFormat + " " + format + "\n"
	fmt.Fprintf(os.Stderr, format, a...)
}

func logInterface(a ...interface{}) {
	log("%v", a...)
}

// Expect structure
type Expect struct {
	Stderr   string
	Stdout   string
	ExitCode int
}

// Assert tests if command executed exactly as described by Expect
func (e *Expect) Assert(t *testing.T, cmd *cmdrunner.Command) {
	if cmd.ExitCode != e.ExitCode {
		t.Errorf("Wanted exit code %v, got %v instead", e.ExitCode, cmd.ExitCode)
	}

	errString := cmd.Stderr.String()
	outString := cmd.Stdout.String()

	if !stringlib.Similar(errString, e.Stderr) {
		t.Errorf("Wanted Stderr %v, got %v instead", e.Stderr, errString)
	}

	if !stringlib.Similar(outString, e.Stdout) {
		t.Errorf("Wanted Stdout %v, got %v instead", e.Stdout, outString)
	}
}
