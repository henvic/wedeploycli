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
	"time"

	"github.com/wedeploy/cli/color"
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
		panic(errors.New("Invalid channel name."))
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
